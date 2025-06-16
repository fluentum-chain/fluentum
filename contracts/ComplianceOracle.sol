// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";
import "./interfaces/IComplianceVerifier.sol";
import "./interfaces/IAMLProvider.sol";
import "./libraries/ComplianceMath.sol";

contract ComplianceOracle is ReentrancyGuard, Ownable {
    // Constants
    uint256 public constant MAX_RISK_SCORE = 100;
    uint256 public constant MIN_CONFIDENCE = 80;
    uint256 public constant MAX_AMOUNT_THRESHOLD = 1000000 * 10**18; // 1M tokens
    
    // State
    IComplianceVerifier public immutable verifier;
    IAMLProvider public immutable amlProvider;
    AggregatorV3Interface public immutable priceFeed;
    mapping(address => bool) public whitelistedAddresses;
    mapping(address => uint256) public lastScreeningTimestamp;
    mapping(bytes32 => bool) public screenedTransactions;
    
    // Events
    event TransactionScreened(
        bytes32 indexed txId,
        address indexed user,
        address indexed counterparty,
        uint256 amount,
        bool sanctionsMatch,
        bool pepMatch,
        uint256 riskScore,
        uint256 confidence
    );
    event AddressWhitelisted(address indexed address_);
    event AddressRemoved(address indexed address_);
    event AMLProviderUpdated(address indexed provider);
    event PriceFeedUpdated(address indexed feed);
    
    struct ScreeningResult {
        bool sanctionsMatch;
        bool pepMatch;
        uint256 riskScore;
        uint256 confidence;
        bytes32 txId;
    }
    
    constructor(
        address _verifier,
        address _amlProvider,
        address _priceFeed
    ) {
        require(_verifier != address(0), "Invalid verifier");
        require(_amlProvider != address(0), "Invalid AML provider");
        require(_priceFeed != address(0), "Invalid price feed");
        
        verifier = IComplianceVerifier(_verifier);
        amlProvider = IAMLProvider(_amlProvider);
        priceFeed = AggregatorV3Interface(_priceFeed);
    }
    
    function screenTransaction(
        address user,
        uint256 amount,
        address counterparty
    ) external nonReentrant returns (ScreeningResult memory) {
        require(user != address(0), "Invalid user");
        require(counterparty != address(0), "Invalid counterparty");
        require(amount > 0, "Invalid amount");
        
        // Check whitelist
        if (whitelistedAddresses[user]) {
            return ScreeningResult({
                sanctionsMatch: false,
                pepMatch: false,
                riskScore: 0,
                confidence: 100,
                txId: bytes32(0)
            });
        }
        
        // Generate transaction ID
        bytes32 txId = keccak256(
            abi.encodePacked(
                user,
                amount,
                counterparty,
                block.chainid,
                block.timestamp
            )
        );
        
        require(!screenedTransactions[txId], "Transaction already screened");
        screenedTransactions[txId] = true;
        
        // Get USD value
        uint256 usdValue = ComplianceMath.calculateUSDValue(
            amount,
            getLatestPrice()
        );
        
        // Check amount threshold
        if (usdValue > MAX_AMOUNT_THRESHOLD) {
            return ScreeningResult({
                sanctionsMatch: true,
                pepMatch: true,
                riskScore: MAX_RISK_SCORE,
                confidence: 100,
                txId: txId
            });
        }
        
        // Call off-chain screening
        bytes memory result = _callOffchainScreening(
            user,
            amount,
            counterparty,
            usdValue
        );
        
        // Decode result
        (
            bool sanctionsMatch,
            bool pepMatch,
            uint256 riskScore,
            uint256 confidence
        ) = abi.decode(result, (bool, bool, uint256, uint256));
        
        // Verify result
        require(
            verifier.verifyScreening(
                user,
                amount,
                counterparty,
                sanctionsMatch,
                pepMatch,
                riskScore,
                confidence
            ),
            "Invalid screening result"
        );
        
        // Update last screening timestamp
        lastScreeningTimestamp[user] = block.timestamp;
        
        emit TransactionScreened(
            txId,
            user,
            counterparty,
            amount,
            sanctionsMatch,
            pepMatch,
            riskScore,
            confidence
        );
        
        return ScreeningResult({
            sanctionsMatch: sanctionsMatch,
            pepMatch: pepMatch,
            riskScore: riskScore,
            confidence: confidence,
            txId: txId
        });
    }
    
    function _callOffchainScreening(
        address user,
        uint256 amount,
        address counterparty,
        uint256 usdValue
    ) internal returns (bytes memory) {
        return amlProvider.screenTransaction(
            user,
            amount,
            counterparty,
            usdValue
        );
    }
    
    function getLatestPrice() public view returns (uint256) {
        (
            uint80 roundId,
            int256 price,
            uint256 startedAt,
            uint256 updatedAt,
            uint80 answeredInRound
        ) = priceFeed.latestRoundData();
        
        require(price > 0, "Invalid price");
        require(updatedAt > 0, "Round not complete");
        require(answeredInRound >= roundId, "Stale price");
        
        return uint256(price);
    }
    
    function whitelistAddress(address address_) external onlyOwner {
        require(address_ != address(0), "Invalid address");
        require(!whitelistedAddresses[address_], "Already whitelisted");
        
        whitelistedAddresses[address_] = true;
        
        emit AddressWhitelisted(address_);
    }
    
    function removeFromWhitelist(address address_) external onlyOwner {
        require(whitelistedAddresses[address_], "Not whitelisted");
        
        whitelistedAddresses[address_] = false;
        
        emit AddressRemoved(address_);
    }
    
    function updateAMLProvider(address _amlProvider) external onlyOwner {
        require(_amlProvider != address(0), "Invalid provider");
        
        amlProvider = IAMLProvider(_amlProvider);
        
        emit AMLProviderUpdated(_amlProvider);
    }
    
    function updatePriceFeed(address _priceFeed) external onlyOwner {
        require(_priceFeed != address(0), "Invalid feed");
        
        priceFeed = AggregatorV3Interface(_priceFeed);
        
        emit PriceFeedUpdated(_priceFeed);
    }
    
    function isTransactionScreened(
        bytes32 txId
    ) external view returns (bool) {
        return screenedTransactions[txId];
    }
    
    function getLastScreeningTimestamp(
        address user
    ) external view returns (uint256) {
        return lastScreeningTimestamp[user];
    }
} 