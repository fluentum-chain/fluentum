![FLUMX Logo](./src/assets/flumx-logo.png)

# Fluentum Web Wallet

A browser-based self-custody wallet for the Fluentum blockchain, built with React, Vite, and the Fluentum SDK.

## Features
- Connect with MetaMask, WalletConnect, or native Fluentum wallets
- View balances and send FLUMX tokens
- Secure, client-side encrypted key management
- Session timeout and auto-logout
- Responsive, Material UI-based design

## Quick Start

### 1. Clone and Install

```bash
cd fluentum-web-wallet
./scripts/setup.sh
```

Or, if you prefer npm:
```bash
npm install
```

### 2. Configure Environment

Edit `.env` with your Fluentum RPC, chain ID, WalletConnect project ID, and encryption key:

```
VITE_RPC_URL=https://rpc.fluentum.io
VITE_CHAIN_ID=118
VITE_WALLETCONNECT_ID=your_project_id
VITE_STORAGE_KEY=your_encryption_key
```

### 3. Run Locally

```bash
npm run dev
```

### 4. Build and Deploy

```bash
npm run build
npm install -g vercel
vercel --prod
```

## Mobile Support
Add these meta tags to `index.html` for mobile browser support:
```html
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1" />
<meta name="mobile-web-app-capable" content="yes">
```

## Security Checklist
- All private keys are encrypted before storage
- Session auto-logout after inactivity
- CSP headers recommended in production
- Audit dependencies regularly

## License
MIT © Fluentum Technologies 