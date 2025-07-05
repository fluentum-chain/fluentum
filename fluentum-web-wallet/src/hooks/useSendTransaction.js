import { useMutation } from '@tanstack/react-query'
import { fluentum } from '../lib/fluentum'

export function useSendTransaction() {
  return useMutation({
    mutationFn: async ({ to, amount }) => {
      const txHash = await fluentum.sendTransaction({
        to,
        value: fluentum.utils.parseUnits(amount, 18)
      })
      return txHash
    }
  })
} 