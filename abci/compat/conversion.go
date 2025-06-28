package compat

import (
	cosmosproto "cosmossdk.io/api/tendermint/crypto"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	cmcryptoed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmcryptosecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	localabci "github.com/fluentum-chain/fluentum/abci/types"
)

// ToCmPublicKey converts a Cosmos SDK API proto PublicKey to the upstream CometBFT crypto PublicKey.
func ToCmPublicKey(pk *cosmosproto.PublicKey) cmcrypto.PubKey {
	if pk == nil {
		return nil
	}
	if ed := pk.GetEd25519(); ed != nil {
		var pubKey cmcryptoed25519.PubKey
		copy(pubKey[:], ed)
		return pubKey
	}
	if secp := pk.GetSecp256K1(); secp != nil {
		var pubKey cmcryptosecp256k1.PubKey
		copy(pubKey[:], secp)
		return pubKey
	}
	return nil
}

// ToLocalPublicKey converts an upstream CometBFT crypto PublicKey to the Cosmos SDK API proto PublicKey.
func ToLocalPublicKey(pk cmcrypto.PubKey) *cosmosproto.PublicKey {
	if pk == nil {
		return &cosmosproto.PublicKey{}
	}

	// For now, we'll return a simple implementation
	// In a real implementation, you'd need to handle different key types properly
	return &cosmosproto.PublicKey{
		Sum: &cosmosproto.PublicKey_Ed25519{Ed25519: pk.Bytes()},
	}
}

// ExecTxResult conversion
func ExecTxResultFromComet(src *cmtabci.ExecTxResult) *localabci.ExecTxResult {
	if src == nil {
		return nil
	}
	return &localabci.ExecTxResult{
		Code:      src.Code,
		Data:      src.Data,
		Log:       src.Log,
		Info:      src.Info,
		GasWanted: src.GasWanted,
		GasUsed:   src.GasUsed,
		Events:    nil, // TODO: Convert events if needed
		Codespace: src.Codespace,
	}
}

// CheckTxResponse conversion
func CheckTxResponseFromComet(src *cmtabci.ResponseCheckTx) *localabci.CheckTxResponse {
	if src == nil {
		return nil
	}
	return &localabci.CheckTxResponse{
		Code:      src.Code,
		Data:      src.Data,
		Log:       src.Log,
		Info:      src.Info,
		GasWanted: src.GasWanted,
		GasUsed:   src.GasUsed,
		Events:    nil, // TODO: Convert events if needed
		Codespace: src.Codespace,
	}
}

// ConsensusParams conversion
func ConsensusParamsFromComet(src *cmttypes.ConsensusParams) *localabci.ConsensusParams {
	if src == nil {
		return nil
	}
	return &localabci.ConsensusParams{
		Block:     nil, // TODO: Convert BlockParams if needed
		Evidence:  nil, // TODO: Convert EvidenceParams if needed
		Validator: nil, // TODO: Convert ValidatorParams if needed
		Version:   nil, // TODO: Convert VersionParams if needed
	}
}

// Convert slice of ExecTxResult
func ExecTxResultsFromComet(src []*cmtabci.ExecTxResult) []*localabci.ExecTxResult {
	if src == nil {
		return nil
	}
	out := make([]*localabci.ExecTxResult, len(src))
	for i, txr := range src {
		out[i] = ExecTxResultFromComet(txr)
	}
	return out
}

// ProofOps conversion
func ProofOpsFromComet(src *cmcrypto.ProofOps) *localabci.ProofOps {
	if src == nil {
		return nil
	}
	// Convert from CometBFT ProofOps to local ProofOps
	ops := make([]*localabci.ProofOp, len(src.Ops))
	for i, op := range src.Ops {
		ops[i] = &localabci.ProofOp{
			Type: op.Type,
			Key:  op.Key,
			Data: op.Data,
		}
	}
	return &localabci.ProofOps{
		Ops: ops,
	}
}
