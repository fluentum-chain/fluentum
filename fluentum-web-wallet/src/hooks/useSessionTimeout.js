import { useEffect } from 'react'
import { useWallet } from '../contexts/WalletContext'

export function useSessionTimeout(timeout = 30 * 60 * 1000) {
  const { disconnect } = useWallet()

  useEffect(() => {
    let timer
    const resetTimer = () => {
      clearTimeout(timer)
      timer = setTimeout(disconnect, timeout)
    }

    window.addEventListener('mousemove', resetTimer)
    window.addEventListener('keypress', resetTimer)
    resetTimer()

    return () => {
      clearTimeout(timer)
      window.removeEventListener('mousemove', resetTimer)
      window.removeEventListener('keypress', resetTimer)
    }
  }, [disconnect, timeout])
} 