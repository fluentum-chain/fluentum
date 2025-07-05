import { createContext, useContext, useState } from 'react'
import { fluentum } from '../lib/fluentum'

const WalletContext = createContext()

export function WalletProvider({ children }) {
  const [address, setAddress] = useState('')
  const [balance, setBalance] = useState('0')
  const [connector, setConnector] = useState(null)

  const connectWallet = async (type) => {
    try {
      const { address, connector } = await fluentum.connect(type)
      setAddress(address)
      setConnector(connector)
      await updateBalance(address)
    } catch (error) {
      console.error('Connection error:', error)
    }
  }

  const updateBalance = async (addr) => {
    const bal = await fluentum.getBalance(addr)
    setBalance(bal)
  }

  return (
    <WalletContext.Provider value={{ address, balance, connectWallet }}>
      {children}
    </WalletContext.Provider>
  )
}

export const useWallet = () => useContext(WalletContext) 