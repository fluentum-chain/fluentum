import { useState } from 'react'
import { TextField, Button } from '@mui/material'
import { useSendTransaction } from '../hooks/useSendTransaction'

export function SendForm() {
  const [recipient, setRecipient] = useState('')
  const [amount, setAmount] = useState('')
  const { mutate, isLoading } = useSendTransaction()

  const handleSubmit = (e) => {
    e.preventDefault()
    mutate({ to: recipient, amount })
  }

  return (
    <form onSubmit={handleSubmit}>
      <TextField
        label="Recipient"
        value={recipient}
        onChange={(e) => setRecipient(e.target.value)}
        fullWidth
        margin="normal"
      />
      <TextField
        label="Amount (FLUMX)"
        type="number"
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
        fullWidth
        margin="normal"
      />
      <Button 
        type="submit" 
        variant="contained" 
        disabled={isLoading}
      >
        {isLoading ? 'Sending...' : 'Send'}
      </Button>
    </form>
  )
} 