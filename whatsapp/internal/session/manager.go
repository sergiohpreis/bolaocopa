// PROTOTYPE — throwaway. Answers: does whatsmeow QR+session+group-link flow work for bolaocopa?
package session

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "github.com/mattn/go-sqlite3"
)

type State string

const (
	StateDisconnected State = "disconnected"
	StateConnecting   State = "connecting"
	StateAwaitingQR   State = "awaiting_qr"
	StateConnected    State = "connected"
)

type Group struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

type Manager struct {
	mu          sync.RWMutex
	client      *whatsmeow.Client
	state       State
	qrBase64    string
	linkedGroup string // JID do grupo vinculado
	enabled     bool   // notificações automáticas ativas
	storePath   string
}

func New(storePath string) *Manager {
	m := &Manager{
		state:     StateDisconnected,
		storePath: storePath,
	}
	m.linkedGroup = m.loadLinkedGroup()
	m.enabled = m.loadEnabled()
	return m
}

func (m *Manager) linkedGroupFile() string {
	return filepath.Join(m.storePath, "linked_group")
}

func (m *Manager) enabledFile() string {
	return filepath.Join(m.storePath, "enabled")
}

func (m *Manager) loadLinkedGroup() string {
	b, err := os.ReadFile(m.linkedGroupFile())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func (m *Manager) loadEnabled() bool {
	b, err := os.ReadFile(m.enabledFile())
	if err != nil {
		return true // default: enabled
	}
	return strings.TrimSpace(string(b)) == "true"
}

func (m *Manager) saveLinkedGroup(jid string) {
	if err := os.WriteFile(m.linkedGroupFile(), []byte(jid), 0600); err != nil {
		slog.Error("persist linked group", "err", err)
	}
}

func (m *Manager) saveEnabled(v bool) {
	val := "false"
	if v {
		val = "true"
	}
	if err := os.WriteFile(m.enabledFile(), []byte(val), 0600); err != nil {
		slog.Error("persist enabled flag", "err", err)
	}
}

func (m *Manager) resetToDisconnected() {
	m.mu.Lock()
	m.state = StateDisconnected
	m.mu.Unlock()
}

func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()

	// Allow reconnect from disconnected or expired-QR states
	if m.state != StateDisconnected && m.state != StateAwaitingQR {
		m.mu.Unlock()
		return fmt.Errorf("already %s", m.state)
	}

	// Tear down any existing client before reconnecting
	if m.client != nil {
		m.client.Disconnect()
		m.client = nil
	}
	// Hold the slot before releasing the lock to prevent TOCTOU races
	m.state = StateConnecting
	m.qrBase64 = ""
	m.mu.Unlock()

	container, err := sqlstore.New(ctx, "sqlite3", "file:"+m.storePath+"/whatsapp.db?_foreign_keys=on", waLog.Noop)
	if err != nil {
		m.resetToDisconnected()
		return fmt.Errorf("sqlstore: %w", err)
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		m.resetToDisconnected()
		return fmt.Errorf("get device: %w", err)
	}

	client := whatsmeow.NewClient(device, waLog.Noop)
	client.AddEventHandler(m.handleEvent)

	if client.Store.ID == nil {
		// New device: need QR
		qrChan, err := client.GetQRChannel(ctx)
		if err != nil {
			m.resetToDisconnected()
			return fmt.Errorf("get qr channel: %w", err)
		}
		if err := client.Connect(); err != nil {
			m.resetToDisconnected()
			return fmt.Errorf("connect: %w", err)
		}

		m.mu.Lock()
		m.client = client
		m.state = StateAwaitingQR
		m.mu.Unlock()

		go func() {
			for evt := range qrChan {
				switch evt.Event {
				case "code":
					png, encErr := qrcode.Encode(evt.Code, qrcode.Medium, 256)
					if encErr != nil {
						slog.Error("qrcode encode", "err", encErr)
						continue
					}
					m.mu.Lock()
					m.qrBase64 = base64.StdEncoding.EncodeToString(png)
					m.mu.Unlock()
					slog.Info("QR code updated")
				case "timeout":
					// QR expired without being scanned — back to disconnected
					m.mu.Lock()
					m.state = StateDisconnected
					m.client = nil
					m.qrBase64 = ""
					m.mu.Unlock()
					slog.Warn("QR code expired, disconnected")
				}
			}
		}()
	} else {
		// Reconnect existing session
		if err := client.Connect(); err != nil {
			m.resetToDisconnected()
			return fmt.Errorf("reconnect: %w", err)
		}
		m.mu.Lock()
		m.client = client
		m.state = StateConnected
		m.mu.Unlock()
		slog.Info("reconnected existing WhatsApp session")
	}

	return nil
}

func (m *Manager) handleEvent(evt interface{}) {
	switch evt.(type) {
	case *events.Connected:
		m.mu.Lock()
		m.state = StateConnected
		m.qrBase64 = ""
		m.mu.Unlock()
		slog.Info("WhatsApp connected")
	case *events.Disconnected:
		// Do NOT nil m.client here — whatsmeow reconnects automatically and
		// fires *events.Connected again reusing the same client instance.
		// Niling would break ListGroups/SendText during the reconnect window.
		m.mu.Lock()
		m.state = StateDisconnected
		m.mu.Unlock()
		slog.Warn("WhatsApp disconnected")
	case *events.LoggedOut:
		// LoggedOut is terminal — the client will not reconnect on its own.
		m.mu.Lock()
		m.state = StateDisconnected
		m.client = nil
		m.mu.Unlock()
		slog.Warn("WhatsApp logged out")
	}
}

func (m *Manager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Manager) QRBase64() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.qrBase64
}

func (m *Manager) LinkedGroup() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.linkedGroup
}

func (m *Manager) LinkGroup(jid string) {
	m.mu.Lock()
	m.linkedGroup = jid
	m.mu.Unlock()
	m.saveLinkedGroup(jid)
}

func (m *Manager) Enabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

func (m *Manager) SetEnabled(v bool) {
	m.mu.Lock()
	m.enabled = v
	m.mu.Unlock()
	m.saveEnabled(v)
}

func (m *Manager) ListGroups(ctx context.Context) ([]Group, error) {
	m.mu.RLock()
	client := m.client
	state := m.state
	m.mu.RUnlock()

	if state != StateConnected || client == nil {
		return nil, fmt.Errorf("not connected")
	}

	groups, err := client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}

	result := make([]Group, 0, len(groups))
	for _, g := range groups {
		result = append(result, Group{JID: g.JID.String(), Name: g.Name})
	}
	return result, nil
}

func (m *Manager) SendText(ctx context.Context, jid, text string) error {
	m.mu.RLock()
	client := m.client
	state := m.state
	m.mu.RUnlock()

	if state != StateConnected || client == nil {
		return fmt.Errorf("not connected")
	}

	recipient, err := types.ParseJID(jid)
	if err != nil {
		return fmt.Errorf("parse jid: %w", err)
	}

	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	_, err = client.SendMessage(ctx, recipient, msg)
	return err
}

func (m *Manager) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		m.client.Disconnect()
		m.client = nil
	}
	m.state = StateDisconnected
}
