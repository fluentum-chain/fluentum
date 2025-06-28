package compat

import (
	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	cmcryptoed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmcryptosecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	localabci "github.com/fluentum-chain/fluentum/abci/types"
	protoabci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
	protocrypto "github.com/fluentum-chain/fluentum/proto/tendermint/crypto"
	prototypes "github.com/fluentum-chain/fluentum/proto/tendermint/types"
)

// ToCmPublicKey converts a local proto PublicKey to the upstream CometBFT crypto PublicKey.
func ToCmPublicKey(pk *protocrypto.PublicKey) cmcrypto.PubKey {
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
func ToLocalPublicKey(pk cmcrypto.PubKey) *protocrypto.PublicKey {
	if pk == nil {
		return &protocrypto.PublicKey{}
	}

	// For now, we'll return a simple implementation
	// In a real implementation, you'd need to handle different key types properly
	return &protocrypto.PublicKey{
		Sum: &protocrypto.PublicKey_Ed25519{Ed25519: pk.Bytes()},
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
func ConsensusParamsFromComet(src *cmttypes.ConsensusParams) *protoabci.ConsensusParams {
	if src == nil {
		return nil
	}
	return &protoabci.ConsensusParams{
		Block:     &protoabci.BlockParams{},      // Use abci package types
		Evidence:  &prototypes.EvidenceParams{},  // Use types1 alias
		Validator: &prototypes.ValidatorParams{}, // Use types1 alias
		Version:   &prototypes.VersionParams{},   // Use types1 alias
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
func ProofOpsFromComet(src *cmtcrypto.ProofOps) *cmtcrypto.ProofOps {
	if src == nil {
		return nil
	}
	// Since we're using CometBFT crypto types, just return the same type
	return src
}

// ConsensusParamsToProtoTypes converts abci.ConsensusParams to types.ConsensusParams
func ConsensusParamsToProtoTypes(src *protoabci.ConsensusParams) *prototypes.ConsensusParams {
	if src == nil {
		return nil
	}

	var block prototypes.BlockParams
	if src.Block != nil {
		block = prototypes.BlockParams{
			MaxBytes: src.Block.MaxBytes,
			MaxGas:   src.Block.MaxGas,
		}
	}

	var evidence prototypes.EvidenceParams
	if src.Evidence != nil {
		evidence = prototypes.EvidenceParams{
			MaxAgeNumBlocks: src.Evidence.MaxAgeNumBlocks,
			MaxAgeDuration:  src.Evidence.MaxAgeDuration,
			MaxBytes:        src.Evidence.MaxBytes,
		}
	}

	var validator prototypes.ValidatorParams
	if src.Validator != nil {
		validator = prototypes.ValidatorParams{
			PubKeyTypes: src.Validator.PubKeyTypes,
		}
	}

	var version prototypes.VersionParams
	if src.Version != nil {
		version = prototypes.VersionParams{
			AppVersion: src.Version.AppVersion,
		}
	}

	return &prototypes.ConsensusParams{
		Block:     block,
		Evidence:  evidence,
		Validator: validator,
		Version:   version,
	}
}
