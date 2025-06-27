package abci

import (
	"fmt"
	"reflect"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/abci/types"
)

// VerifyCompliance checks that all implementations comply with the ABCI interface
func VerifyCompliance() error {
	// Check MyApp compliance
	var app types.Application = &MyApp{}

	// Get the Application interface type
	appType := reflect.TypeOf((*types.Application)(nil)).Elem()

	// Get the MyApp type
	myAppType := reflect.TypeOf(app)

	// Check if MyApp implements all methods of Application
	if !myAppType.Implements(appType) {
		return fmt.Errorf("MyApp does not implement the Application interface")
	}

	// Check BaseApplication compliance
	var baseApp types.Application = &types.BaseApplication{}
	baseAppType := reflect.TypeOf(baseApp)

	if !baseAppType.Implements(appType) {
		return fmt.Errorf("BaseApplication does not implement the Application interface")
	}

	// Verify all required methods exist
	requiredMethods := []string{
		"Echo",
		"Info",
		"Query",
		"CheckTx",
		"PrepareProposal",
		"ProcessProposal",
		"FinalizeBlock",
		"ExtendVote",
		"VerifyVoteExtension",
		"Commit",
		"InitChain",
		"ListSnapshots",
		"OfferSnapshot",
		"LoadSnapshotChunk",
		"ApplySnapshotChunk",
	}

	for _, methodName := range requiredMethods {
		_, found := appType.MethodByName(methodName)
		if !found {
			return fmt.Errorf("Application interface missing required method: %s", methodName)
		}

		_, found = myAppType.MethodByName(methodName)
		if !found {
			return fmt.Errorf("MyApp missing required method: %s", methodName)
		}

		_, found = baseAppType.MethodByName(methodName)
		if !found {
			return fmt.Errorf("BaseApplication missing required method: %s", methodName)
		}
	}

	fmt.Println("✓ All implementations comply with the ABCI interface")
	return nil
}

// VerifyTypeCompatibility checks that all types are compatible with CometBFT
func VerifyTypeCompatibility() error {
	// Check that response codes are properly defined
	if types.CodeTypeOK != cmtabci.CodeTypeOK {
		return fmt.Errorf("CodeTypeOK should be %d, got %d", cmtabci.CodeTypeOK, types.CodeTypeOK)
	}

	// Check helper functions
	if !types.IsOK(types.CodeTypeOK) {
		return fmt.Errorf("IsOK should return true for CodeTypeOK")
	}

	if types.IsOK(types.CodeTypeInternalError) {
		return fmt.Errorf("IsOK should return false for CodeTypeInternalError")
	}

	if !types.IsError(types.CodeTypeInternalError) {
		return fmt.Errorf("IsError should return true for CodeTypeInternalError")
	}

	if types.IsError(types.CodeTypeOK) {
		return fmt.Errorf("IsError should return false for CodeTypeOK")
	}

	fmt.Println("✓ All types are compatible with CometBFT")
	return nil
}

// RunAllVerifications runs all verification checks
func RunAllVerifications() error {
	fmt.Println("Running ABCI interface compliance verifications...")

	if err := VerifyCompliance(); err != nil {
		return fmt.Errorf("compliance verification failed: %v", err)
	}

	if err := VerifyTypeCompatibility(); err != nil {
		return fmt.Errorf("type compatibility verification failed: %v", err)
	}

	fmt.Println("✓ All verifications passed!")
	return nil
}
