package com.fluentum.wallet.utils

import org.fluentum.sdk.crypto.Bech32

class AddressValidator {
    companion object {
        private val FLUENTUM_ADDRESS_REGEX = Regex("^fluentum1[a-zA-Z0-9]{38}$")
        
        fun isValidFluentumAddress(address: String): Boolean {
            return try {
                FLUENTUM_ADDRESS_REGEX.matches(address) && 
                Bech32.decode(address).first == "fluentum"
            } catch (e: Exception) {
                false
            }
        }
        
        fun formatAddress(address: String): String {
            return if (address.length > 20) {
                "${address.take(10)}...${address.takeLast(10)}"
            } else {
                address
            }
        }
        
        fun validateAndFormat(address: String): String? {
            return if (isValidFluentumAddress(address)) {
                formatAddress(address)
            } else {
                null
            }
        }
    }
} 