import { Dialog, Button, Box } from '@mui/material'
import QRCode from 'qrcode.react'
import { useWallet } from '../contexts/WalletContext'
import { fluentum } from '../lib/fluentum'

export function ConnectModal({ open, onClose }) {
  const { connectWallet } = useWallet()

  return (
    <Dialog open={open} onClose={onClose}>
      <Box sx={{ p: 4 }}>
        <Button 
          variant="contained" 
          onClick={() => connectWallet('injected')}
          sx={{ mb: 2 }}
        >
          MetaMask
        </Button>
        <Button 
          variant="contained" 
          onClick={() => connectWallet('walletconnect')}
        >
          WalletConnect
        </Button>
        <Box sx={{ mt: 4 }}>
          <QRCode 
            value={fluentum.getWalletConnectUri()} 
            size={256}
          />
        </Box>
      </Box>
    </Dialog>
  )
} 