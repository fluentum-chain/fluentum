package com.fluentum.wallet.network

import org.fluentum.sdk.FluentumSDK
import org.fluentum.sdk.config.FluentumConfig

enum class FluentumNetwork(
    val chainId: String, 
    val rpcUrl: String, 
    val restUrl: String,
    val explorerUrl: String,
    val displayName: String
) {
    MAINNET(
        "fluentum-1",
        "https://rpc.fluentum.io",
        "https://api.fluentum.io",
        "https://explorer.fluentum.io",
        "Fluentum Mainnet"
    ),
    TESTNET(
        "fluentum-testnet-1",
        "https://testnet-rpc.fluentum.io",
        "https://testnet-api.fluentum.io",
        "https://testnet-explorer.fluentum.io",
        "Fluentum Testnet"
    ),
    DEVNET(
        "fluentum-devnet-1",
        "https://devnet-rpc.fluentum.io",
        "https://devnet-api.fluentum.io",
        "https://devnet-explorer.fluentum.io",
        "Fluentum Devnet"
    ),
    LOCAL(
        "fluentum-local",
        "http://localhost:26657",
        "http://localhost:1317",
        "http://localhost:8080",
        "Local Network"
    );
    
    companion object {
        fun fromChainId(chainId: String): FluentumNetwork? {
            return values().find { it.chainId == chainId }
        }
    }
}

class NetworkManager {
    private var currentNetwork = FluentumNetwork.TESTNET
    
    fun getCurrentNetwork(): FluentumNetwork = currentNetwork
    
    fun switchNetwork(network: FluentumNetwork) {
        currentNetwork = network
        FluentumSDK.updateConfig(
            FluentumConfig(
                chainId = network.chainId,
                rpcEndpoint = network.rpcUrl,
                restEndpoint = network.restUrl
            )
        )
    }
    
    fun isMainnet(): Boolean = currentNetwork == FluentumNetwork.MAINNET
    
    fun isTestnet(): Boolean = currentNetwork == FluentumNetwork.TESTNET
    
    fun getExplorerUrl(): String = currentNetwork.explorerUrl
    
    fun getRpcUrl(): String = currentNetwork.rpcUrl
    
    fun getRestUrl(): String = currentNetwork.restUrl
} 