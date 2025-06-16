// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IStaking {
    function getDelegationInfo(uint256 validatorId, address delegator) external view returns (uint256 amount, uint256 pendingRewards);
    function getValidatorInfo(uint256 validatorId) external view returns (
        address owner,
        uint256 stakedAmount,
        uint256 commissionRate,
        uint256 totalDelegated,
        bool active,
        uint256 activationEpoch
    );
    function delegate(uint256 validatorId, uint256 amount) external;
    function undelegate(uint256 validatorId, uint256 amount) external;
    function claimRewards(uint256 validatorId) external;
    function createValidator(uint256 amount, uint256 commission) external;
    function deactivateValidator(uint256 validatorId) external;
} 