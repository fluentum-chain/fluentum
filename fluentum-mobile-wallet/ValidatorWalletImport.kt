package com.fluentum.wallet.import

import android.util.Base64
import com.fluentum.wallet.backup.Wallet
import com.fluentum.wallet.utils.AddressValidator
import org.fluentum.sdk.FluentumSDK
import org.fluentum.sdk.crypto.KeyPair

class ValidatorWalletImport {
    
    fun importValidatorKey(privateKeyBase64: String): Wallet {
        try {
            val privateKeyBytes = Base64.decode(privateKeyBase64, Base64.DEFAULT)
            val keyPair = FluentumSDK.importFromPrivateKey(privateKeyBytes)
            
            return Wallet(
                address = keyPair.address,
                publicKey = keyPair.publicKey,
                privateKey = keyPair.privateKey
            )
        } catch (e: Exception) {
            throw IllegalArgumentException("Invalid private key format: ${e.message}")
        }
    }
    
    fun importValidatorKeyFromHex(privateKeyHex: String): Wallet {
        try {
            val privateKeyBytes = privateKeyHex.chunked(2).map { it.toInt(16).toByte() }.toByteArray()
            val keyPair = FluentumSDK.importFromPrivateKey(privateKeyBytes)
            
            return Wallet(
                address = keyPair.address,
                publicKey = keyPair.publicKey,
                privateKey = keyPair.privateKey
            )
        } catch (e: Exception) {
            throw IllegalArgumentException("Invalid hex private key: ${e.message}")
        }
    }
    
    fun validateValidatorAddress(address: String): Boolean {
        return AddressValidator.isValidFluentumAddress(address)
    }
    
    fun getValidatorAddressFromPrivateKey(privateKeyBase64: String): String {
        val wallet = importValidatorKey(privateKeyBase64)
        return wallet.address
    }
    
    fun createValidatorWallet(
        privateKeyBase64: String,
        validatorName: String = "Validator"
    ): ValidatorWallet {
        val wallet = importValidatorKey(privateKeyBase64)
        return ValidatorWallet(
            wallet = wallet,
            name = validatorName,
            isValidator = true
        )
    }
    
    companion object {
        // Your validator's private key from priv_validator_key.json
        const val VALIDATOR_PRIVATE_KEY = "N7DB9x6XIjaTq0cjVJjszrGGHXRRY6N3jrTnSQye8If5mDLiRrb0E16ziCuGORjSZi7aUCfUG+G3GmMsrKyWTw=="
        
        // Your validator's address
        const val VALIDATOR_ADDRESS = "fluentum1repgx0ynpf7kjr2cmmptv483w3xhvpslrzsc0q"
    }
}

data class ValidatorWallet(
    val wallet: Wallet,
    val name: String,
    val isValidator: Boolean,
    val stakingAmount: Long = 0,
    val commission: Double = 0.0
) {
    val address: String get() = wallet.address
    val formattedStakingAmount: String get() = "${stakingAmount / 1_000_000.0} FLUMX"
    val formattedCommission: String get() = "${commission * 100}%"
} 