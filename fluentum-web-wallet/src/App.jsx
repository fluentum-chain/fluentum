import { useState } from 'react'
import { WalletProvider, useWallet } from './contexts/WalletContext'
import { ConnectModal } from './components/ConnectModal'
import { BalanceCard } from './components/BalanceCard'
import { SendForm } from './components/SendForm'
import { Button } from '@mui/material'

function App() {
  const [modalOpen, setModalOpen] = useState(false)
  const { address } = useWallet()

  return (
    <div className="app">
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