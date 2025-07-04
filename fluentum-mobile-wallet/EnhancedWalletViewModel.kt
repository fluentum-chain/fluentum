package com.fluentum.wallet.viewmodel

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.fluentum.wallet.backup.Wallet
import com.fluentum.wallet.backup.WalletBackup
import com.fluentum.wallet.data.Transaction
import com.fluentum.wallet.data.TransactionRepository
import com.fluentum.wallet.data.WalletResult
import com.fluentum.wallet.import.ValidatorWalletImport
import com.fluentum.wallet.network.NetworkManager
import com.fluentum.wallet.utils.AddressValidator
import kotlinx.coroutines.launch
import org.fluentum.sdk.FluentumSDK
import org.fluentum.sdk.model.BalanceResponse

class EnhancedWalletViewModel : ViewModel() {
    
    private val walletBackup = WalletBackup(context)
    private val transactionRepository = TransactionRepository()
    private val networkManager = NetworkManager()
    private val validatorImport = ValidatorWalletImport()
    
    // LiveData for UI updates
    private val _walletState = MutableLiveData<WalletResult<Wallet>>()
    val walletState: LiveData<WalletResult<Wallet>> = _walletState
    
    private val _balanceState = MutableLiveData<WalletResult<String>>()
    val balanceState: LiveData<WalletResult<String>> = _balanceState
    
    private val _transactionState = MutableLiveData<WalletResult<List<Transaction>>>()
    val transactionState: LiveData<WalletResult<List<Transaction>>> = _transactionState
    
    private val _networkState = MutableLiveData<NetworkManager>()
    val networkState: LiveData<NetworkManager> = _networkState
    
    private val _validatorWalletState = MutableLiveData<WalletResult<ValidatorWallet>>()
    val validatorWalletState: LiveData<WalletResult<ValidatorWallet>> = _validatorWalletState
    
    // Current wallet
    private var currentWallet: Wallet? = null
    
    init {
        _networkState.value = networkManager
    }
    
    // Wallet Management
    fun createNewWallet() {
        viewModelScope.launch {
            _walletState.value = WalletResult.Loading
            try {
                val wallet = walletBackup.generateNewWallet()
                currentWallet = wallet
                _walletState.value = WalletResult.Success(wallet)
                fetchBalance(wallet.address)
            } catch (e: Exception) {
                _walletState.value = WalletResult.Error("Failed to create wallet: ${e.message}")
            }
        }
    }
    
    fun importFromMnemonic(mnemonic: String) {
        viewModelScope.launch {
            _walletState.value = WalletResult.Loading
            try {
                if (!walletBackup.validateMnemonic(mnemonic)) {
                    _walletState.value = WalletResult.Error("Invalid mnemonic phrase")
                    return@launch
                }
                
                val wallet = walletBackup.importFromMnemonic(mnemonic)
                currentWallet = wallet
                _walletState.value = WalletResult.Success(wallet)
                fetchBalance(wallet.address)
            } catch (e: Exception) {
                _walletState.value = WalletResult.Error("Failed to import wallet: ${e.message}")
            }
        }
    }
    
    fun importValidatorWallet() {
        viewModelScope.launch {
            _validatorWalletState.value = WalletResult.Loading
            try {
                val validatorWallet = validatorImport.createValidatorWallet(
                    ValidatorWalletImport.VALIDATOR_PRIVATE_KEY,
                    "Fluentum Validator"
                )
                currentWallet = validatorWallet.wallet
                _validatorWalletState.value = WalletResult.Success(validatorWallet)
                fetchBalance(validatorWallet.address)
            } catch (e: Exception) {
                _validatorWalletState.value = WalletResult.Error("Failed to import validator wallet: ${e.message}")
            }
        }
    }
    
    // Balance Management
    fun fetchBalance(address: String) {
        viewModelScope.launch {
            _balanceState.value = WalletResult.Loading
            try {
                val response = FluentumSDK.queryBalance(address)
                val balance = formatBalance(response)
                _balanceState.value = WalletResult.Success(balance)
            } catch (e: Exception) {
                _balanceState.value = WalletResult.Error("Failed to fetch balance: ${e.message}")
            }
        }
    }
    
    private fun formatBalance(response: BalanceResponse): String {
        val flumxBalance = response.balances.find { it.denom == "uflumx" }
        return if (flumxBalance != null) {
            "${flumxBalance.amount.toLong() / 1_000_000.0} FLUMX"
        } else {
            "0 FLUMX"
        }
    }
    
    // Transaction Management
    fun sendTokens(
        senderAddress: String,
        recipientAddress: String,
        amount: Long,
        memo: String = ""
    ) {
        viewModelScope.launch {
            _transactionState.value = WalletResult.Loading
            
            try {
                // Validate addresses
                if (!AddressValidator.isValidFluentumAddress(recipientAddress)) {
                    _transactionState.value = WalletResult.Error("Invalid recipient address")
                    return@launch
                }
                
                // Validate amount
                if (amount <= 0) {
                    _transactionState.value = WalletResult.Error("Invalid amount")
                    return@launch
                }
                
                val wallet = currentWallet ?: throw IllegalStateException("No wallet loaded")
                val privateKey = wallet.privateKey ?: throw IllegalStateException("No private key available")
                
                // Create and sign transaction
                val tx = FluentumSDK.createTransferTx(
                    senderAddress,
                    recipientAddress,
                    amount,
                    "uflumx",
                    memo
                )
                
                val signedTx = FluentumSDK.signTransaction(tx, privateKey)
                val result = FluentumSDK.broadcastTransaction(signedTx)
                
                if (result.isSuccess) {
                    _transactionState.value = WalletResult.Success(emptyList())
                    // Refresh balance
                    fetchBalance(senderAddress)
                } else {
                    _transactionState.value = WalletResult.Error("Transaction failed: ${result.errorMessage}")
                }
                
            } catch (e: Exception) {
                _transactionState.value = WalletResult.Error("Failed to send tokens: ${e.message}")
            }
        }
    }
    
    fun fetchTransactionHistory(address: String) {
        viewModelScope.launch {
            _transactionState.value = WalletResult.Loading
            try {
                val transactions = transactionRepository.getTransactionHistory(address)
                _transactionState.value = WalletResult.Success(transactions)
            } catch (e: Exception) {
                _transactionState.value = WalletResult.Error("Failed to fetch transaction history: ${e.message}")
            }
        }
    }
    
    // Network Management
    fun switchNetwork(network: FluentumNetwork) {
        viewModelScope.launch {
            try {
                networkManager.switchNetwork(network)
                _networkState.value = networkManager
                
                // Refresh data for new network
                currentWallet?.let { fetchBalance(it.address) }
            } catch (e: Exception) {
                // Handle network switch error
            }
        }
    }
    
    // Backup and Recovery
    fun exportMnemonic(): String? {
        val wallet = currentWallet ?: return null
        return try {
            walletBackup.exportMnemonic(wallet)
        } catch (e: Exception) {
            null
        }
    }
    
    fun exportKeystore(password: String): String? {
        val wallet = currentWallet ?: return null
        return try {
            walletBackup.exportKeystore(wallet, password)
        } catch (e: Exception) {
            null
        }
    }
    
    // Utility functions
    fun getCurrentWallet(): Wallet? = currentWallet
    
    fun getCurrentNetwork(): FluentumNetwork = networkManager.getCurrentNetwork()
    
    fun isMainnet(): Boolean = networkManager.isMainnet()
    
    fun isTestnet(): Boolean = networkManager.isTestnet()
    
    fun validateAddress(address: String): Boolean = AddressValidator.isValidFluentumAddress(address)
    
    fun formatAddress(address: String): String = AddressValidator.formatAddress(address)
} 