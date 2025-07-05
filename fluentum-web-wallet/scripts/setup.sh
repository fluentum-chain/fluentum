#!/bin/bash

# Install Vite and React tooling
yarn add -D vite @vitejs/plugin-react || npm install --save-dev vite @vitejs/plugin-react

# Install core dependencies
yarn add @fluentum-web/sdk ethers @tanstack/react-query @metamask/providers @walletconnect/web3-provider qrcode.react @mui/material @emotion/react @emotion/styled framer-motion crypto-js || npm install @fluentum-web/sdk ethers @tanstack/react-query @metamask/providers @walletconnect/web3-provider qrcode.react @mui/material @emotion/react @emotion/styled framer-motion crypto-js

# Success message
echo "✅ All Fluentum web wallet dependencies installed." 