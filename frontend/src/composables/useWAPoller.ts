// PROTOTYPE — throwaway. Encapsulates the WhatsApp status polling interval.
import { onUnmounted } from 'vue'

export function useWAPoller(callback: () => void, intervalMs = 5000) {
  const timer = setInterval(callback, intervalMs)

  onUnmounted(() => clearInterval(timer))

  return {
    stop: () => clearInterval(timer),
  }
}
