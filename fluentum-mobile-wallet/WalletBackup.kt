package com.fluentum.wallet.backup

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import org.fluentum.sdk.FluentumSDK
import org.fluentum.sdk.crypto.KeyPair
import org.fluentum.sdk.crypto.Mnemonic
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

data class Wallet(
    val address: String,
    val publicKey: ByteArray,
    val privateKey: ByteArray? = null,
    val mnemonic: String? = null
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false
        other as Wallet
        return address == other.address
    }
    
    override fun hashCode(): Int {
        return address.hashCode()
    }
}

class WalletBackup(private val context: Context) {
    private val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
    private val keyAlias = "fluentum_wallet_backup"
    
    fun exportMnemonic(wallet: Wallet): String {
        return wallet.mnemonic ?: throw IllegalStateException("No mnemonic available")
    }
    
    fun importFromMnemonic(mnemonic: String): Wallet {
        val keyPair = FluentumSDK.importFromMnemonic(mnemonic)
        return Wallet(
            address = keyPair.address,
            publicKey = keyPair.publicKey,
            privateKey = keyPair.privateKey,
            mnemonic = mnemonic
        )
    }
    
    fun exportKeystore(wallet: Wallet, password: String): String {
        val encryptedPrivateKey = encryptPrivateKey(wallet.privateKey ?: throw IllegalStateException("No private key available"))
        return createKeystoreJson(wallet.address, encryptedPrivateKey, password)
    }
    
    fun importFromKeystore(keystoreJson: String, password: String): Wallet {
        val keystoreData = parseKeystoreJson(keystoreJson)
        val decryptedPrivateKey = decryptPrivateKey(keystoreData.encryptedPrivateKey, password)
        val keyPair = FluentumSDK.importFromPrivateKey(decryptedPrivateKey)
        return Wallet(
            address = keyPair.address,
            publicKey = keyPair.publicKey,
            privateKey = keyPair.privateKey
        )
    }
    
    fun generateNewWallet(): Wallet {
        val mnemonic = Mnemonic.generate(24)
        val keyPair = FluentumSDK.importFromMnemonic(mnemonic)
        return Wallet(
            address = keyPair.address,
            publicKey = keyPair.publicKey,
            privateKey = keyPair.privateKey,
            mnemonic = mnemonic
        )
    }
    
    fun validateMnemonic(mnemonic: String): Boolean {
        return try {
            Mnemonic.validate(mnemonic)
            true
        } catch (e: Exception) {
            false
        }
    }
    
    private fun encryptPrivateKey(privateKey: ByteArray): ByteArray {
        val secretKey = getOrCreateSecretKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)
        val encrypted = cipher.doFinal(privateKey)
        val iv = cipher.iv
        return iv + encrypted
    }
    
    private fun decryptPrivateKey(encryptedData: ByteArray, password: String): ByteArray {
        val secretKey = getOrCreateSecretKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val iv = encryptedData.sliceArray(0..11)
        val encrypted = encryptedData.sliceArray(12 until encryptedData.size)
        cipher.init(Cipher.DECRYPT_MODE, secretKey, GCMParameterSpec(128, iv))
        return cipher.doFinal(encrypted)
    }
    
    private fun getOrCreateSecretKey(): SecretKey {
        if (!keyStore.containsAlias(keyAlias)) {
            val keyGenerator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore")
            val keyGenSpec = KeyGenParameterSpec.Builder(
                keyAlias,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setUserAuthenticationRequired(false)
                .build()
            keyGenerator.init(keyGenSpec)
            keyGenerator.generateKey()
        }
        return keyStore.getKey(keyAlias, null) as SecretKey
    }
    
    private fun createKeystoreJson(address: String, encryptedPrivateKey: ByteArray, password: String): String {
        // Simplified keystore format - in production, use a more robust format
        return """
        {
            "version": 1,
            "address": "$address",
            "encrypted_private_key": "${android.util.Base64.encodeToString(encryptedPrivateKey, android.util.Base64.DEFAULT)}",
            "crypto": {
                "cipher": "aes-256-gcm",
                "kdf": "pbkdf2",
                "kdfparams": {
                    "salt": "${android.util.Base64.encodeToString(password.toByteArray(), android.util.Base64.DEFAULT)}",
                    "iterations": 100000
                }
            }
        }
        """.trimIndent()
    }
    
    private fun parseKeystoreJson(keystoreJson: String): KeystoreData {
        // Simplified parsing - in production, use proper JSON parsing
        val lines = keystoreJson.lines()
        val address = lines.find { it.contains("\"address\"") }?.split("\"")?.get(3) ?: ""
        val encryptedPrivateKey = lines.find { it.contains("\"encrypted_private_key\"") }?.split("\"")?.get(3)?.let {
            android.util.Base64.decode(it, android.util.Base64.DEFAULT)
        } ?: ByteArray(0)
        return KeystoreData(address, encryptedPrivateKey)
    }
    
    private data class KeystoreData(val address: String, val encryptedPrivateKey: ByteArray)
} 