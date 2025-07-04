package com.fluentum.wallet.security

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import java.io.File
import java.security.MessageDigest

class IntegrityChecker(private val context: Context) {
    
    companion object {
        private const val EXPECTED_APP_SIGNATURE = "your_expected_app_signature_here"
        private const val EXPECTED_PACKAGE_NAME = "com.fluentum.wallet"
    }
    
    fun checkAppIntegrity(): IntegrityResult {
        val checks = mutableListOf<IntegrityCheck>()
        
        // Check if app is debuggable
        val isDebuggable = checkIfDebuggable()
        checks.add(IntegrityCheck("Debug Mode", !isDebuggable, isDebuggable))
        
        // Check if running on emulator
        val isEmulator = checkIfEmulator()
        checks.add(IntegrityCheck("Emulator Detection", !isEmulator, isEmulator))
        
        // Check app signature
        val isValidSignature = checkAppSignature()
        checks.add(IntegrityCheck("App Signature", isValidSignature, !isValidSignature))
        
        // Check package name
        val isValidPackage = checkPackageName()
        checks.add(IntegrityCheck("Package Name", isValidPackage, !isValidPackage))
        
        // Check for root
        val isRooted = checkIfRooted()
        checks.add(IntegrityCheck("Root Detection", !isRooted, isRooted))
        
        // Check for suspicious files
        val hasSuspiciousFiles = checkForSuspiciousFiles()
        checks.add(IntegrityCheck("Suspicious Files", !hasSuspiciousFiles, hasSuspiciousFiles))
        
        // Check for hooking frameworks
        val hasHookingFrameworks = checkForHookingFrameworks()
        checks.add(IntegrityCheck("Hooking Frameworks", !hasHookingFrameworks, hasHookingFrameworks))
        
        val allChecksPassed = checks.all { it.passed }
        
        return IntegrityResult(
            isSecure = allChecksPassed,
            checks = checks,
            riskLevel = calculateRiskLevel(checks)
        )
    }
    
    private fun checkIfDebuggable(): Boolean {
        return try {
            val applicationInfo = context.applicationInfo
            (applicationInfo.flags and android.content.pm.ApplicationInfo.FLAG_DEBUGGABLE) != 0
        } catch (e: Exception) {
            false
        }
    }
    
    private fun checkIfEmulator(): Boolean {
        return try {
            val buildModel = Build.MODEL.lowercase()
            val buildManufacturer = Build.MANUFACTURER.lowercase()
            val buildProduct = Build.PRODUCT.lowercase()
            val buildFingerprint = Build.FINGERPRINT.lowercase()
            
            val emulatorIndicators = listOf(
                "sdk", "google_sdk", "sdk_x86", "vbox86p", "vbox86tp", "emulator",
                "android sdk built for x86", "generic", "generic_x86", "generic_x86_64",
                "sdk_gphone", "sdk_gphone_x86", "sdk_gphone_x86_64", "sdk_gphone64_x86_64",
                "sdk_gphone64_arm64", "sdk_gphone_armv7", "sdk_gphone_armv8", "sdk_gphone_armv8_64",
                "sdk_gphone_armv8_32", "sdk_gphone_armv7_64", "sdk_gphone_armv7_32",
                "sdk_gphone_armv6", "sdk_gphone_armv6_64", "sdk_gphone_armv6_32",
                "sdk_gphone_armv5", "sdk_gphone_armv5_64", "sdk_gphone_armv5_32"
            )
            
            emulatorIndicators.any { indicator ->
                buildModel.contains(indicator) ||
                buildManufacturer.contains(indicator) ||
                buildProduct.contains(indicator) ||
                buildFingerprint.contains(indicator)
            }
        } catch (e: Exception) {
            false
        }
    }
    
    private fun checkAppSignature(): Boolean {
        return try {
            val packageInfo = context.packageManager.getPackageInfo(
                context.packageName,
                PackageManager.GET_SIGNATURES
            )
            
            val signatures = packageInfo.signatures
            if (signatures.isNotEmpty()) {
                val signature = signatures[0]
                val signatureHash = hashSignature(signature.toByteArray())
                signatureHash == EXPECTED_APP_SIGNATURE
            } else {
                false
            }
        } catch (e: Exception) {
            false
        }
    }
    
    private fun checkPackageName(): Boolean {
        return context.packageName == EXPECTED_PACKAGE_NAME
    }
    
    private fun checkIfRooted(): Boolean {
        val rootIndicators = listOf(
            "/system/app/Superuser.apk",
            "/sbin/su",
            "/system/bin/su",
            "/system/xbin/su",
            "/data/local/xbin/su",
            "/data/local/bin/su",
            "/system/sd/xbin/su",
            "/system/bin/failsafe/su",
            "/data/local/su",
            "/su/bin/su"
        )
        
        return rootIndicators.any { File(it).exists() }
    }
    
    private fun checkForSuspiciousFiles(): Boolean {
        val suspiciousFiles = listOf(
            "/system/bin/su",
            "/system/xbin/su",
            "/sbin/su",
            "/system/app/Superuser.apk",
            "/system/etc/init.d/99SuperSUDaemon",
            "/dev/com.koushikdutta.superuser.daemon/"
        )
        
        return suspiciousFiles.any { File(it).exists() }
    }
    
    private fun checkForHookingFrameworks(): Boolean {
        val hookingFrameworks = listOf(
            "com.saurik.substrate",
            "de.robv.android.xposed.installer",
            "com.topjohnwu.magisk",
            "com.kingroot.kinguser",
            "com.noshufou.android.su",
            "com.thirdparty.superuser",
            "eu.chainfire.supersu",
            "com.noxx.substratum"
        )
        
        return hookingFrameworks.any { packageName ->
            try {
                context.packageManager.getPackageInfo(packageName, 0)
                true
            } catch (e: PackageManager.NameNotFoundException) {
                false
            }
        }
    }
    
    private fun hashSignature(signature: ByteArray): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val hash = digest.digest(signature)
        return hash.joinToString("") { "%02x".format(it) }
    }
    
    private fun calculateRiskLevel(checks: List<IntegrityCheck>): RiskLevel {
        val failedChecks = checks.count { !it.passed }
        
        return when {
            failedChecks == 0 -> RiskLevel.LOW
            failedChecks <= 2 -> RiskLevel.MEDIUM
            failedChecks <= 4 -> RiskLevel.HIGH
            else -> RiskLevel.CRITICAL
        }
    }
    
    fun getDeviceInfo(): DeviceInfo {
        return DeviceInfo(
            model = Build.MODEL,
            manufacturer = Build.MANUFACTURER,
            product = Build.PRODUCT,
            fingerprint = Build.FINGERPRINT,
            sdkVersion = Build.VERSION.SDK_INT.toString(),
            androidVersion = Build.VERSION.RELEASE,
            isEmulator = checkIfEmulator(),
            isRooted = checkIfRooted(),
            isDebuggable = checkIfDebuggable()
        )
    }
}

data class IntegrityResult(
    val isSecure: Boolean,
    val checks: List<IntegrityCheck>,
    val riskLevel: RiskLevel
) {
    val failedChecks: List<IntegrityCheck> get() = checks.filter { !it.passed }
    val passedChecks: List<IntegrityCheck> get() = checks.filter { it.passed }
    val totalChecks: Int get() = checks.size
    val passedCount: Int get() = passedChecks.size
    val failedCount: Int get() = failedChecks.size
}

data class IntegrityCheck(
    val name: String,
    val passed: Boolean,
    val isSecurityRisk: Boolean
) {
    val status: String get() = if (passed) "PASS" else "FAIL"
    val description: String get() = if (isSecurityRisk) {
        "Security risk detected: $name"
    } else {
        "Integrity check: $name"
    }
}

enum class RiskLevel {
    LOW, MEDIUM, HIGH, CRITICAL;
    
    fun getDescription(): String = when (this) {
        LOW -> "Low risk - App appears secure"
        MEDIUM -> "Medium risk - Some security concerns detected"
        HIGH -> "High risk - Multiple security issues detected"
        CRITICAL -> "Critical risk - App may be compromised"
    }
    
    fun getColor(): Int = when (this) {
        LOW -> 0xFF4CAF50 // Green
        MEDIUM -> 0xFFFF9800 // Orange
        HIGH -> 0xFFFF5722 // Red
        CRITICAL -> 0xFFD32F2F // Dark Red
    }
}

data class DeviceInfo(
    val model: String,
    val manufacturer: String,
    val product: String,
    val fingerprint: String,
    val sdkVersion: String,
    val androidVersion: String,
    val isEmulator: Boolean,
    val isRooted: Boolean,
    val isDebuggable: Boolean
) {
    val deviceType: String get() = when {
        isEmulator -> "Emulator"
        isRooted -> "Rooted Device"
        else -> "Physical Device"
    }
    
    val securityStatus: String get() = when {
        isRooted -> "Compromised"
        isEmulator -> "Test Environment"
        isDebuggable -> "Development Mode"
        else -> "Secure"
    }
} 