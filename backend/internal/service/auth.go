package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/sergiohpreis/bolaocopa/backend/internal/repository"
)

var (
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrEmailExists     = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthService struct {
	q          repository.Querier
	jwtSecret  []byte
	oauthCfg   *oauth2.Config
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type GoogleUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"picture"`
}

func NewAuthService(q repository.Querier, jwtSecret, googleClientID, googleClientSecret, googleRedirectURL string) *AuthService {
	oauthCfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &AuthService{q: q, jwtSecret: []byte(jwtSecret), oauthCfg: oauthCfg}
}

func (s *AuthService) GetGoogleAuthURL(state string) string {
	return s.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *AuthService) ExchangeGoogleCode(ctx context.Context, code string) (TokenResponse, error) {
	token, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("exchanging oauth code: %w", err)
	}

	info, err := fetchGoogleUserInfo(ctx, s.oauthCfg.Client(ctx, token))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("fetching google user info: %w", err)
	}

	user, err := s.q.UpsertUserByGoogleID(ctx, repository.UpsertUserByGoogleIDParams{
		GoogleID:  pgtype.Text{String: info.ID, Valid: true},
		Email:     info.Email,
		Name:      info.Name,
		AvatarUrl: pgtype.Text{String: info.AvatarURL, Valid: info.AvatarURL != ""},
	})
	if err != nil {
		return TokenResponse{}, fmt.Errorf("upserting user: %w", err)
	}

	return s.generateTokenPair(uuidToString(user.ID), user.Name)
}

func (s *AuthService) ValidateAccessToken(tokenString string) (string, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return "", ErrInvalidToken
	}
	typ, _ := claims["typ"].(string)
	if typ != "access" {
		return "", ErrInvalidToken
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", ErrInvalidToken
	}
	return sub, nil
}

func (s *AuthService) ValidateUserExists(ctx context.Context, userID string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return ErrInvalidToken
	}
	if _, err := s.q.GetUserByID(ctx, uid); err != nil {
		return ErrInvalidToken
	}
	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (TokenResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}
	typ, _ := claims["typ"].(string)
	if typ != "refresh" {
		return TokenResponse{}, ErrInvalidToken
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return TokenResponse{}, ErrInvalidToken
	}
	uid, err := parseUUID(sub)
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}
	user, err := s.q.GetUserByID(ctx, uid)
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}
	return s.generateTokenPair(sub, user.Name)
}

func (s *AuthService) generateTokenPair(userID, name string) (TokenResponse, error) {
	accessExp := time.Now().Add(15 * time.Minute)
	accessToken, err := s.signToken(userID, name, "access", accessExp)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("signing access token: %w", err)
	}
	refreshExp := time.Now().Add(7 * 24 * time.Hour)
	refreshToken, err := s.signToken(userID, name, "refresh", refreshExp)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("signing refresh token: %w", err)
	}
	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Until(accessExp).Seconds()),
	}, nil
}

func (s *AuthService) signToken(userID, name, typ string, exp time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"name": name,
		"typ":  typ,
		"exp":  exp.Unix(),
		"iat":  time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *AuthService) RegisterByEmail(ctx context.Context, email, name, password string) (TokenResponse, error) {
	if len(password) < 8 {
		return TokenResponse{}, fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.q.CreateUserByEmail(ctx, repository.CreateUserByEmailParams{
		Email:        email,
		Name:         name,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return TokenResponse{}, ErrEmailExists
		}
		return TokenResponse{}, fmt.Errorf("creating user: %w", err)
	}

	return s.generateTokenPair(uuidToString(user.ID), user.Name)
}

func (s *AuthService) LoginByEmail(ctx context.Context, email, password string) (TokenResponse, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return TokenResponse{}, ErrInvalidCredentials
	} else if err != nil {
		return TokenResponse{}, fmt.Errorf("getting user: %w", err)
	}

	if !user.PasswordHash.Valid {
		return TokenResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	return s.generateTokenPair(uuidToString(user.ID), user.Name)
}

func (s *AuthService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
