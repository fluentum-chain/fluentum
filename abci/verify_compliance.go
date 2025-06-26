package abci

import (
	"fmt"
	"reflect"

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

// VerifyOptionalInterfaces checks that optional interfaces are properly defined
func VerifyOptionalInterfaces() error {
	// Check Snapshotter interface
	snapshotterType := reflect.TypeOf((*types.Snapshotter)(nil)).Elem()
	requiredSnapshotterMethods := []string{
		"ListSnapshots",
		"OfferSnapshot", 
		"LoadSnapshotChunk",
		"ApplySnapshotChunk",
	}
	
	for _, methodName := range requiredSnapshotterMethods {
		_, found := snapshotterType.MethodByName(methodName)
		if !found {
			return fmt.Errorf("Snapshotter interface missing required method: %s", methodName)
		}
	}
	
	// Check ValidatorSetUpdater interface
	validatorUpdaterType := reflect.TypeOf((*types.ValidatorSetUpdater)(nil)).Elem()
	_, found := validatorUpdaterType.MethodByName("ApplyValidatorSetUpdates")
	if !found {
		return fmt.Errorf("ValidatorSetUpdater interface missing required method: ApplyValidatorSetUpdates")
	}
	
	// Check ProposalProcessor interface
	proposalProcessorType := reflect.TypeOf((*types.ProposalProcessor)(nil)).Elem()
	requiredProposalMethods := []string{
		"PrepareProposal",
		"ProcessProposal",
	}
	
	for _, methodName := range requiredProposalMethods {
		_, found := proposalProcessorType.MethodByName(methodName)
		if !found {
			return fmt.Errorf("ProposalProcessor interface missing required method: %s", methodName)
		}
	}
	
	// Check VoteExtensionProcessor interface
	voteExtProcessorType := reflect.TypeOf((*types.VoteExtensionProcessor)(nil)).Elem()
	requiredVoteExtMethods := []string{
		"ExtendVote",
		"VerifyVoteExtension",
	}
	
	for _, methodName := range requiredVoteExtMethods {
		_, found := voteExtProcessorType.MethodByName(methodName)
		if !found {
			return fmt.Errorf("VoteExtensionProcessor interface missing required method: %s", methodName)
		}
	}
	
	fmt.Println("✓ All optional interfaces are properly defined")
	return nil
}

// VerifyTypeCompatibility checks that all types are compatible with CometBFT
func VerifyTypeCompatibility() error {
	// Check that response codes are properly defined
	if types.CodeTypeOK != 0 {
		return fmt.Errorf("CodeTypeOK should be 0, got %d", types.CodeTypeOK)
	}
	
	if types.CodeTypeErr != 1 {
		return fmt.Errorf("CodeTypeErr should be 1, got %d", types.CodeTypeErr)
	}
	
	// Check helper functions
	if !types.IsOK(types.CodeTypeOK) {
		return fmt.Errorf("IsOK should return true for CodeTypeOK")
	}
	
	if types.IsOK(types.CodeTypeErr) {
		return fmt.Errorf("IsOK should return false for CodeTypeErr")
	}
	
	if !types.IsError(types.CodeTypeErr) {
		return fmt.Errorf("IsError should return true for CodeTypeErr")
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
	
	if err := VerifyOptionalInterfaces(); err != nil {
		return fmt.Errorf("optional interfaces verification failed: %v", err)
	}
	
	if err := VerifyTypeCompatibility(); err != nil {
		return fmt.Errorf("type compatibility verification failed: %v", err)
	}
	
	fmt.Println("✓ All verifications passed!")
	return nil
} 