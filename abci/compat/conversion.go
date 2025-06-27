package compat

import (
	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	cmcryptoed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmcryptosecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	localabci "github.com/fluentum-chain/fluentum/abci/types"
	localcrypto "github.com/fluentum-chain/fluentum/proto/tendermint/crypto"
	localtypes "github.com/fluentum-chain/fluentum/proto/tendermint/types"
)

// ToCmPublicKey converts a local proto PublicKey to the upstream CometBFT crypto PublicKey.
func ToCmPublicKey(pk *localcrypto.PublicKey) cmcrypto.PubKey {
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

// ToLocalPublicKey converts an upstream CometBFT crypto PublicKey to the local proto PublicKey.
func ToLocalPublicKey(pk cmcrypto.PubKey) *localcrypto.PublicKey {
	if pk == nil {
		return &localcrypto.PublicKey{}
	}

	// For now, we'll return a simple implementation
	// In a real implementation, you'd need to handle different key types properly
	return &localcrypto.PublicKey{
		Sum: &localcrypto.PublicKey_Ed25519{Ed25519: pk.Bytes()},
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
func ConsensusParamsFromComet(src *cmttypes.ConsensusParams) *localtypes.ConsensusParams {
	if src == nil {
		return nil
	}
	return &localtypes.ConsensusParams{
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
func ProofOpsFromComet(src *cmtcrypto.ProofOps) *localcrypto.ProofOps {
	if src == nil {
		return nil
	}
	// Convert each ProofOp in the slice
	ops := make([]*localcrypto.ProofOp, len(src.Ops))
	for i, op := range src.Ops {
		ops[i] = &localcrypto.ProofOp{
			Type: op.Type,
			Key:  op.Key,
			Data: op.Data,
		}
	}
	return &localcrypto.ProofOps{
		Ops: ops,
	}
}
