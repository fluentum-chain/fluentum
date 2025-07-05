import { Card, Typography } from '@mui/material'
import { useWallet } from '../contexts/WalletContext'

export function BalanceCard() {
  const { balance } = useWallet()

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h5">
        {balance} FLUMX
      </Typography>
    </Card>
  )
} 