package com.fluentum.wallet.ui

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Bundle
import android.view.View
import android.widget.EditText
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import androidx.lifecycle.ViewModelProvider
import com.fluentum.wallet.R
import com.fluentum.wallet.backup.Wallet
import com.fluentum.wallet.data.Transaction
import com.fluentum.wallet.data.WalletResult
import com.fluentum.wallet.import.ValidatorWallet
import com.fluentum.wallet.network.FluentumNetwork
import com.fluentum.wallet.viewmodel.EnhancedWalletViewModel
import com.google.android.material.snackbar.Snackbar
import com.journeyapps.barcodescanner.ScanContract
import com.journeyapps.barcodescanner.ScanOptions
import android.content.ClipData
import android.content.ClipboardManager

class EnhancedMainActivity : AppCompatActivity() {
    
    private lateinit var binding: ActivityMainBinding
    private lateinit var viewModel: EnhancedWalletViewModel
    
    private val requestPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            // Permission granted, can proceed with camera operations
        } else {
            Snackbar.make(binding.root, "Camera permission required for QR scanning", Snackbar.LENGTH_LONG).show()
        }
    }
    
    private val qrCodeLauncher = registerForActivityResult(ScanContract()) { result ->
        result.contents?.let { scannedAddress ->
            binding.etRecipient.setText(scannedAddress)
            if (viewModel.validateAddress(scannedAddress)) {
                Snackbar.make(binding.root, "Valid Fluentum address scanned", Snackbar.LENGTH_SHORT).show()
            } else {
                Snackbar.make(binding.root, "Invalid address format", Snackbar.LENGTH_SHORT).show()
            }
        }
    }
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)
        
        // Initialize ViewModel
        viewModel = ViewModelProvider(this)[EnhancedWalletViewModel::class.java]
        
        setupUI()
        setupObservers()
        setupClickListeners()
        
        // Check if wallet exists, otherwise show setup
        checkWalletSetup()
    }
    
    private fun setupUI() {
        // Set up toolbar
        setSupportActionBar(binding.toolbar)
        supportActionBar?.title = "Fluentum Wallet"
        
        // Set up bottom navigation
        binding.bottomNavigation.setOnItemSelectedListener { item ->
            when (item.itemId) {
                R.id.nav_wallet -> {
                    showWalletFragment()
                    true
                }
                R.id.nav_send -> {
                    showSendFragment()
                    true
                }
                R.id.nav_history -> {
                    showHistoryFragment()
                    true
                }
                R.id.nav_settings -> {
                    showSettingsFragment()
                    true
                }
                else -> false
            }
        }
    }
    
    private fun setupObservers() {
        // Observe wallet state
        viewModel.walletState.observe(this) { result ->
            when (result) {
                is WalletResult.Loading -> {
                    binding.progressBar.visibility = View.VISIBLE
                }
                is WalletResult.Success -> {
                    binding.progressBar.visibility = View.GONE
                    updateWalletUI(result.data)
                }
                is WalletResult.Error -> {
                    binding.progressBar.visibility = View.GONE
                    showError(result.message)
                }
            }
        }
        
        // Observe balance state
        viewModel.balanceState.observe(this) { result ->
            when (result) {
                is WalletResult.Loading -> {
                    binding.tvBalance.text = "Loading..."
                }
                is WalletResult.Success -> {
                    binding.tvBalance.text = result.data
                }
                is WalletResult.Error -> {
                    binding.tvBalance.text = "Error loading balance"
                    showError(result.message)
                }
            }
        }
        
        // Observe transaction state
        viewModel.transactionState.observe(this) { result ->
            when (result) {
                is WalletResult.Loading -> {
                    binding.progressBar.visibility = View.VISIBLE
                }
                is WalletResult.Success -> {
                    binding.progressBar.visibility = View.GONE
                    updateTransactionHistory(result.data)
                }
                is WalletResult.Error -> {
                    binding.progressBar.visibility = View.GONE
                    showError(result.message)
                }
            }
        }
        
        // Observe validator wallet state
        viewModel.validatorWalletState.observe(this) { result ->
            when (result) {
                is WalletResult.Loading -> {
                    binding.progressBar.visibility = View.VISIBLE
                }
                is WalletResult.Success -> {
                    binding.progressBar.visibility = View.GONE
                    updateValidatorWalletUI(result.data)
                }
                is WalletResult.Error -> {
                    binding.progressBar.visibility = View.GONE
                    showError(result.message)
                }
            }
        }
    }
    
    private fun setupClickListeners() {
        // Send button
        binding.btnSend.setOnClickListener {
            sendTokens()
        }
        
        // QR scan button
        binding.btnScanQr.setOnClickListener {
            checkCameraPermissionAndScan()
        }
        
        // Create new wallet button
        binding.btnCreateWallet.setOnClickListener {
            showCreateWalletDialog()
        }
        
        // Import wallet button
        binding.btnImportWallet.setOnClickListener {
            showImportWalletDialog()
        }
        
        // Import validator button
        binding.btnImportValidator.setOnClickListener {
            viewModel.importValidatorWallet()
        }
        
        // Network selector
        binding.spinnerNetwork.setOnItemSelectedListener { _, _, position, _ ->
            val networks = FluentumNetwork.values()
            if (position < networks.size) {
                viewModel.switchNetwork(networks[position])
            }
        }
    }
    
    private fun checkWalletSetup() {
        val currentWallet = viewModel.getCurrentWallet()
        if (currentWallet == null) {
            showSetupDialog()
        } else {
            updateWalletUI(currentWallet)
        }
    }
    
    private fun showSetupDialog() {
        AlertDialog.Builder(this)
            .setTitle("Welcome to Fluentum Wallet")
            .setMessage("Choose how you'd like to set up your wallet:")
            .setPositiveButton("Create New Wallet") { _, _ ->
                viewModel.createNewWallet()
            }
            .setNegativeButton("Import Existing") { _, _ ->
                showImportWalletDialog()
            }
            .setNeutralButton("Import Validator") { _, _ ->
                viewModel.importValidatorWallet()
            }
            .setCancelable(false)
            .show()
    }
    
    private fun showCreateWalletDialog() {
        AlertDialog.Builder(this)
            .setTitle("Create New Wallet")
            .setMessage("This will generate a new wallet with a mnemonic phrase. Make sure to save it securely!")
            .setPositiveButton("Create") { _, _ ->
                viewModel.createNewWallet()
            }
            .setNegativeButton("Cancel", null)
            .show()
    }
    
    private fun showImportWalletDialog() {
        val dialogView = layoutInflater.inflate(R.layout.dialog_import_wallet, null)
        val etMnemonic = dialogView.findViewById<EditText>(R.id.etMnemonic)
        
        AlertDialog.Builder(this)
            .setTitle("Import Wallet")
            .setView(dialogView)
            .setPositiveButton("Import") { _, _ ->
                val mnemonic = etMnemonic.text.toString().trim()
                if (mnemonic.isNotEmpty()) {
                    viewModel.importFromMnemonic(mnemonic)
                } else {
                    showError("Please enter a mnemonic phrase")
                }
            }
            .setNegativeButton("Cancel", null)
            .show()
    }
    
    private fun sendTokens() {
        val recipientAddress = binding.etRecipient.text.toString().trim()
        val amountText = binding.etAmount.text.toString().trim()
        val memo = binding.etMemo.text.toString().trim()
        
        // Validate input
        if (recipientAddress.isEmpty()) {
            showError("Please enter recipient address")
            return
        }
        
        if (!viewModel.validateAddress(recipientAddress)) {
            showError("Invalid recipient address")
            return
        }
        
        if (amountText.isEmpty()) {
            showError("Please enter amount")
            return
        }
        
        val amount = try {
            (amountText.toDouble() * 1_000_000).toLong() // Convert to uflumx
        } catch (e: NumberFormatException) {
            showError("Invalid amount")
            return
        }
        
        if (amount <= 0) {
            showError("Amount must be greater than 0")
            return
        }
        
        val currentWallet = viewModel.getCurrentWallet()
        if (currentWallet == null) {
            showError("No wallet loaded")
            return
        }
        
        // Show confirmation dialog
        AlertDialog.Builder(this)
            .setTitle("Confirm Transaction")
            .setMessage("Send ${amount / 1_000_000.0} FLUMX to ${viewModel.formatAddress(recipientAddress)}?")
            .setPositiveButton("Send") { _, _ ->
                viewModel.sendTokens(currentWallet.address, recipientAddress, amount, memo)
            }
            .setNegativeButton("Cancel", null)
            .show()
    }
    
    private fun checkCameraPermissionAndScan() {
        when {
            ContextCompat.checkSelfPermission(this, Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED -> {
                launchQRScanner()
            }
            shouldShowRequestPermissionRationale(Manifest.permission.CAMERA) -> {
                showCameraPermissionRationale()
            }
            else -> {
                requestPermissionLauncher.launch(Manifest.permission.CAMERA)
            }
        }
    }
    
    private fun launchQRScanner() {
        val options = ScanOptions()
            .setDesiredBarcodeFormats(ScanOptions.QR_CODE)
            .setPrompt("Scan Fluentum Address")
            .setBeepEnabled(false)
            .setBarcodeImageEnabled(true)
        
        qrCodeLauncher.launch(options)
    }
    
    private fun showCameraPermissionRationale() {
        AlertDialog.Builder(this)
            .setTitle("Camera Permission")
            .setMessage("Camera permission is needed to scan QR codes for addresses.")
            .setPositiveButton("Grant Permission") { _, _ ->
                requestPermissionLauncher.launch(Manifest.permission.CAMERA)
            }
            .setNegativeButton("Cancel", null)
            .show()
    }
    
    private fun updateWalletUI(wallet: Wallet) {
        binding.tvWalletAddress.text = viewModel.formatAddress(wallet.address)
        binding.tvWalletAddress.setOnClickListener {
            copyToClipboard(wallet.address)
        }
        
        // Show backup options
        binding.btnBackupMnemonic.setOnClickListener {
            showBackupMnemonicDialog(wallet)
        }
        
        binding.btnBackupKeystore.setOnClickListener {
            showBackupKeystoreDialog(wallet)
        }
        
        // Fetch balance
        viewModel.fetchBalance(wallet.address)
        
        // Fetch transaction history
        viewModel.fetchTransactionHistory(wallet.address)
    }
    
    private fun updateValidatorWalletUI(validatorWallet: ValidatorWallet) {
        updateWalletUI(validatorWallet.wallet)
        
        // Show validator-specific info
        binding.tvValidatorInfo.text = "Validator: ${validatorWallet.name}"
        binding.tvStakingAmount.text = "Staking: ${validatorWallet.formattedStakingAmount}"
        binding.tvCommission.text = "Commission: ${validatorWallet.formattedCommission}"
    }
    
    private fun updateTransactionHistory(transactions: List<Transaction>) {
        // Update transaction list adapter
        val adapter = TransactionAdapter(transactions) { transaction ->
            showTransactionDetails(transaction)
        }
        binding.recyclerViewTransactions.adapter = adapter
    }
    
    private fun showTransactionDetails(transaction: Transaction) {
        AlertDialog.Builder(this)
            .setTitle("Transaction Details")
            .setMessage("""
                Hash: ${transaction.hash}
                From: ${viewModel.formatAddress(transaction.from)}
                To: ${viewModel.formatAddress(transaction.to)}
                Amount: ${transaction.formattedAmount}
                Status: ${transaction.status.getDisplayName()}
                Date: ${transaction.date}
                ${if (transaction.memo.isNotEmpty()) "Memo: ${transaction.memo}" else ""}
            """.trimIndent())
            .setPositiveButton("View on Explorer") { _, _ ->
                openExplorer(transaction.hash)
            }
            .setNegativeButton("Close", null)
            .show()
    }
    
    private fun showBackupMnemonicDialog(wallet: Wallet) {
        val mnemonic = viewModel.exportMnemonic()
        if (mnemonic != null) {
            AlertDialog.Builder(this)
                .setTitle("Backup Mnemonic")
                .setMessage("Write down these 24 words in order and keep them safe:\n\n$mnemonic")
                .setPositiveButton("Copy") { _, _ ->
                    copyToClipboard(mnemonic)
                }
                .setNegativeButton("Close", null)
                .show()
        } else {
            showError("Cannot export mnemonic for this wallet")
        }
    }
    
    private fun showBackupKeystoreDialog(wallet: Wallet) {
        val dialogView = layoutInflater.inflate(R.layout.dialog_backup_keystore, null)
        val etPassword = dialogView.findViewById<EditText>(R.id.etPassword)
        
        AlertDialog.Builder(this)
            .setTitle("Backup Keystore")
            .setView(dialogView)
            .setPositiveButton("Export") { _, _ ->
                val password = etPassword.text.toString()
                if (password.isNotEmpty()) {
                    val keystore = viewModel.exportKeystore(password)
                    if (keystore != null) {
                        copyToClipboard(keystore)
                        showSuccess("Keystore copied to clipboard")
                    } else {
                        showError("Failed to export keystore")
                    }
                } else {
                    showError("Please enter a password")
                }
            }
            .setNegativeButton("Cancel", null)
            .show()
    }
    
    private fun copyToClipboard(text: String) {
        val clipboard = getSystemService(CLIPBOARD_SERVICE) as ClipboardManager
        val clip = ClipData.newPlainText("Fluentum", text)
        clipboard.setPrimaryClip(clip)
        showSuccess("Copied to clipboard")
    }
    
    private fun openExplorer(txHash: String) {
        val explorerUrl = "${viewModel.getCurrentNetwork().explorerUrl}/tx/$txHash"
        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(explorerUrl))
        startActivity(intent)
    }
    
    private fun showError(message: String) {
        Snackbar.make(binding.root, message, Snackbar.LENGTH_LONG).show()
    }
    
    private fun showSuccess(message: String) {
        Snackbar.make(binding.root, message, Snackbar.LENGTH_SHORT).show()
    }
    
    private fun showWalletFragment() {
        // Implementation for wallet fragment
    }
    
    private fun showSendFragment() {
        // Implementation for send fragment
    }
    
    private fun showHistoryFragment() {
        // Implementation for history fragment
    }
    
    private fun showSettingsFragment() {
        // Implementation for settings fragment
    }
} 