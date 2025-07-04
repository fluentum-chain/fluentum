package com.fluentum.wallet.security

import android.content.Context
import android.content.SharedPreferences
import androidx.biometric.BiometricManager
import androidx.biometric.BiometricPrompt
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity
import java.security.MessageDigest
import java.util.concurrent.Executor

class SecurityManager(private val context: Context) {
    
    private val prefs: SharedPreferences = context.getSharedPreferences("fluentum_security", Context.MODE_PRIVATE)
    private val executor: Executor = ContextCompat.getMainExecutor(context)
    
    companion object {
        private const val KEY_APP_LOCK_ENABLED = "app_lock_enabled"
        private const val KEY_BIOMETRIC_ENABLED = "biometric_enabled"
        private const val KEY_TRANSACTION_CONFIRMATION_ENABLED = "tx_confirmation_enabled"
        private const val KEY_MAX_TRANSACTION_AMOUNT = "max_tx_amount"
        private const val KEY_FAILED_ATTEMPTS = "failed_attempts"
        private const val KEY_LOCKOUT_UNTIL = "lockout_until"
        private const val MAX_FAILED_ATTEMPTS = 5
        private const val LOCKOUT_DURATION = 30 * 60 * 1000L // 30 minutes
    }
    
    // App Lock
    fun enableAppLock(enabled: Boolean) {
        prefs.edit().putBoolean(KEY_APP_LOCK_ENABLED, enabled).apply()
    }
    
    fun isAppLockEnabled(): Boolean {
        return prefs.getBoolean(KEY_APP_LOCK_ENABLED, false)
    }
    
    // Biometric Authentication
    fun enableBiometric(enabled: Boolean) {
        prefs.edit().putBoolean(KEY_BIOMETRIC_ENABLED, enabled).apply()
    }
    
    fun isBiometricEnabled(): Boolean {
        return prefs.getBoolean(KEY_BIOMETRIC_ENABLED, false)
    }
    
    fun isBiometricAvailable(): Boolean {
        val biometricManager = BiometricManager.from(context)
        return biometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG) == BiometricManager.BIOMETRIC_SUCCESS
    }
    
    fun showBiometricPrompt(
        activity: FragmentActivity,
        onSuccess: () -> Unit,
        onError: (String) -> Unit
    ) {
        val biometricPrompt = BiometricPrompt(activity, executor, object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) {
                onSuccess()
            }
            
            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                onError(errString.toString())
            }
            
            override fun onAuthenticationFailed() {
                recordFailedAttempt()
                onError("Authentication failed")
            }
        })
        
        val promptInfo = BiometricPrompt.PromptInfo.Builder()
            .setTitle("Fluentum Wallet Authentication")
            .setSubtitle("Authenticate to access your wallet")
            .setNegativeButtonText("Use Password")
            .build()
        
        biometricPrompt.authenticate(promptInfo)
    }
    
    // Transaction Confirmation
    fun enableTransactionConfirmation(enabled: Boolean) {
        prefs.edit().putBoolean(KEY_TRANSACTION_CONFIRMATION_ENABLED, enabled).apply()
    }
    
    fun isTransactionConfirmationEnabled(): Boolean {
        return prefs.getBoolean(KEY_TRANSACTION_CONFIRMATION_ENABLED, true)
    }
    
    fun setMaxTransactionAmount(amount: Long) {
        prefs.edit().putLong(KEY_MAX_TRANSACTION_AMOUNT, amount).apply()
    }
    
    fun getMaxTransactionAmount(): Long {
        return prefs.getLong(KEY_MAX_TRANSACTION_AMOUNT, 1000000000L) // 1000 FLUMX default
    }
    
    fun requiresConfirmation(amount: Long): Boolean {
        return isTransactionConfirmationEnabled() && amount > getMaxTransactionAmount()
    }
    
    // Failed Attempt Tracking
    private fun recordFailedAttempt() {
        val currentAttempts = prefs.getInt(KEY_FAILED_ATTEMPTS, 0) + 1
        prefs.edit().putInt(KEY_FAILED_ATTEMPTS, currentAttempts).apply()
        
        if (currentAttempts >= MAX_FAILED_ATTEMPTS) {
            val lockoutUntil = System.currentTimeMillis() + LOCKOUT_DURATION
            prefs.edit().putLong(KEY_LOCKOUT_UNTIL, lockoutUntil).apply()
        }
    }
    
    fun isLockedOut(): Boolean {
        val lockoutUntil = prefs.getLong(KEY_LOCKOUT_UNTIL, 0)
        return System.currentTimeMillis() < lockoutUntil
    }
    
    fun getLockoutRemainingTime(): Long {
        val lockoutUntil = prefs.getLong(KEY_LOCKOUT_UNTIL, 0)
        return maxOf(0, lockoutUntil - System.currentTimeMillis())
    }
    
    fun resetFailedAttempts() {
        prefs.edit()
            .putInt(KEY_FAILED_ATTEMPTS, 0)
            .putLong(KEY_LOCKOUT_UNTIL, 0)
            .apply()
    }
    
    // Address Book
    fun addToAddressBook(name: String, address: String) {
        val addressBook = getAddressBook().toMutableMap()
        addressBook[name] = address
        saveAddressBook(addressBook)
    }
    
    fun removeFromAddressBook(name: String) {
        val addressBook = getAddressBook().toMutableMap()
        addressBook.remove(name)
        saveAddressBook(addressBook)
    }
    
    fun getAddressBook(): Map<String, String> {
        val addressBookString = prefs.getString("address_book", "{}")
        return try {
            // Simple JSON parsing - in production, use proper JSON library
            parseAddressBook(addressBookString ?: "{}")
        } catch (e: Exception) {
            emptyMap()
        }
    }
    
    private fun saveAddressBook(addressBook: Map<String, String>) {
        val addressBookString = addressBook.entries.joinToString(",") { "${it.key}:${it.value}" }
        prefs.edit().putString("address_book", addressBookString).apply()
    }
    
    private fun parseAddressBook(addressBookString: String): Map<String, String> {
        if (addressBookString == "{}") return emptyMap()
        
        return addressBookString.split(",").associate { entry ->
            val parts = entry.split(":")
            if (parts.size == 2) {
                parts[0] to parts[1]
            } else {
                "" to ""
            }
        }.filter { it.key.isNotEmpty() }
    }
    
    // Security Utilities
    fun hashString(input: String): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val hash = digest.digest(input.toByteArray())
        return hash.joinToString("") { "%02x".format(it) }
    }
    
    fun validatePassword(password: String): Boolean {
        // In production, implement proper password validation
        return password.length >= 8 && 
               password.any { it.isUpperCase() } &&
               password.any { it.isLowerCase() } &&
               password.any { it.isDigit() }
    }
    
    fun generateSecurePassword(): String {
        val chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"
        return (1..16).map { chars.random() }.joinToString("")
    }
    
    // Security Status
    fun getSecurityStatus(): SecurityStatus {
        return SecurityStatus(
            appLockEnabled = isAppLockEnabled(),
            biometricEnabled = isBiometricEnabled(),
            biometricAvailable = isBiometricAvailable(),
            transactionConfirmationEnabled = isTransactionConfirmationEnabled(),
            maxTransactionAmount = getMaxTransactionAmount(),
            failedAttempts = prefs.getInt(KEY_FAILED_ATTEMPTS, 0),
            isLockedOut = isLockedOut(),
            lockoutRemainingTime = getLockoutRemainingTime()
        )
    }
}

data class SecurityStatus(
    val appLockEnabled: Boolean,
    val biometricEnabled: Boolean,
    val biometricAvailable: Boolean,
    val transactionConfirmationEnabled: Boolean,
    val maxTransactionAmount: Long,
    val failedAttempts: Int,
    val isLockedOut: Boolean,
    val lockoutRemainingTime: Long
) {
    val isSecure: Boolean get() = appLockEnabled || biometricEnabled
    val formattedMaxAmount: String get() = "${maxTransactionAmount / 1_000_000.0} FLUMX"
    val formattedLockoutTime: String get() = "${lockoutRemainingTime / 60000} minutes"
} 