package com.fluentum.wallet.data

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.fluentum.sdk.FluentumSDK
import org.fluentum.sdk.model.TransactionResponse

class TransactionRepository {
    
    suspend fun getTransactionHistory(address: String): List<Transaction> = withContext(Dispatchers.IO) {
        try {
            val response = FluentumSDK.queryTransactions(address)
            response.map { it.toTransaction() }
        } catch (e: Exception) {
            emptyList()
        }
    }
    
    suspend fun getTransactionByHash(hash: String): Transaction? = withContext(Dispatchers.IO) {
        try {
            val response = FluentumSDK.queryTransaction(hash)
            response?.toTransaction()
        } catch (e: Exception) {
            null
        }
    }
    
    suspend fun getPendingTransactions(address: String): List<Transaction> = withContext(Dispatchers.IO) {
        try {
            val response = FluentumSDK.queryPendingTransactions(address)
            response.map { it.toTransaction() }
        } catch (e: Exception) {
            emptyList()
        }
    }
    
    private fun TransactionResponse.toTransaction(): Transaction {
        return Transaction(
            hash = this.hash,
            from = this.from,
            to = this.to,
            amount = this.amount,
            denom = this.denom,
            timestamp = this.timestamp,
            status = when (this.status) {
                "pending" -> TransactionStatus.PENDING
                "confirmed" -> TransactionStatus.CONFIRMED
                "failed" -> TransactionStatus.FAILED
                else -> TransactionStatus.UNKNOWN
            },
            fee = this.fee,
            memo = this.memo ?: "",
            blockHeight = this.blockHeight
        )
    }
} 