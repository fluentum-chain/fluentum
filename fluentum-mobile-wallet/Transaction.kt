package com.fluentum.wallet.data

import java.util.Date

data class Transaction(
    val hash: String,
    val from: String,
    val to: String,
    val amount: Long,
    val denom: String,
    val timestamp: Long,
    val status: TransactionStatus,
    val fee: Long = 0,
    val memo: String = "",
    val blockHeight: Long? = null
) {
    val date: Date get() = Date(timestamp)
    
    val formattedAmount: String get() = "${amount / 1_000_000.0} FLUMX"
    val formattedFee: String get() = "${fee / 1_000_000.0} FLUMX"
}

enum class TransactionStatus {
    PENDING, CONFIRMED, FAILED, UNKNOWN;
    
    fun getDisplayName(): String = when (this) {
        PENDING -> "Pending"
        CONFIRMED -> "Confirmed"
        FAILED -> "Failed"
        UNKNOWN -> "Unknown"
    }
    
    fun getColor(): Int = when (this) {
        PENDING -> 0xFFFFA500 // Orange
        CONFIRMED -> 0xFF4CAF50 // Green
        FAILED -> 0xFFF44336 // Red
        UNKNOWN -> 0xFF9E9E9E // Gray
    }
} 