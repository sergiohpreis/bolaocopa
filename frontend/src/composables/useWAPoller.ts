// PROTOTYPE — throwaway. Encapsulates the WhatsApp status polling interval.
import { onScopeDispose } from 'vue'

export function useWAPoller(callback: () => void, intervalMs = 5000) {
  const timer = setInterval(callback, intervalMs)

  onScopeDispose(() => clearInterval(timer))

  return {
    stop: () => clearInterval(timer),
  }
}
