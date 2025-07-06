import { useState } from 'react'
import { WalletProvider, useWallet } from './contexts/WalletContext'
import { ConnectModal } from './components/ConnectModal'
import { BalanceCard } from './components/BalanceCard'
import { SendForm } from './components/SendForm'
import { Button } from '@mui/material'
import flumxLogo from './assets/flumx-logo.png'

function App() {
  const [modalOpen, setModalOpen] = useState(false)
  const { address } = useWallet()

  return (
    <div className="app">
      <img src={flumxLogo} alt="FLUMX Logo" style={{ width: 120, margin: '32px auto', display: 'block' }} />
      {!address ? (
        <Button onClick={() => setModalOpen(true)}>Connect Wallet</Button>
      ) : (
        <>
          <BalanceCard />
          <SendForm />
        </>
      )}
      <ConnectModal open={modalOpen} onClose={() => setModalOpen(false)} />
    </div>
  )
}

export default function Root() {
  return (
    <WalletProvider>
      <App />
    </WalletProvider>
  )
} 