package com.fluentum.wallet.data

sealed class WalletResult<out T> {
    data class Success<T>(val data: T) : WalletResult<T>()
    data class Error(val message: String, val code: Int? = null, val exception: Exception? = null) : WalletResult<Nothing>()
    object Loading : WalletResult<Nothing>()
    
    fun isSuccess(): Boolean = this is Success
    fun isError(): Boolean = this is Error
    fun isLoading(): Boolean = this is Loading
    
    fun getOrNull(): T? = when (this) {
        is Success -> data
        else -> null
    }
    
    fun getOrThrow(): T = when (this) {
        is Success -> data
        is Error -> throw exception ?: RuntimeException(message)
        is Loading -> throw IllegalStateException("Result is still loading")
    }
} 