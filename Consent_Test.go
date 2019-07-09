package consent

import (
	"chaincode/solution_template/test_utils"
	"common/bchcls/global_data"
	common_test_utils "common/bchcls/test_utils"
	"common/bchcls/user_mgmt"
	"encoding/json"
	//"reflect"
	"testing"
	//"time"
)

// Sets up indices and registers caller
func setup(t *testing.T) (*common_test_utils.NewMockStub, global_data.User) {
	// Setup indices
	stub := test_utils.SetupIndexesAndGetStub(t)
	stub.MockTransactionStart("t1")
	SetupIndex(stub)
	stub.MockTransactionEnd("t1")

	// Register caller
	caller := common_test_utils.CreateTestUser("caller")
	stub.MockTransactionStart("t1")
	user_mgmt.RegisterUserInternal(stub, caller, caller, false)
	stub.MockTransactionEnd("t1")

	return stub, caller
}

func TestPutConsent(t *testing.T) {
	// Setup indices & register caller
	stub, caller := setup(t)

	// Create myConsent
	myConsent := Consent{
		ConsentID: "E3", 
		Name: "Satya Majumder", 
		Email: "satya@gmail.com", 
		Phone: 6127022297,
		EventName: "JPMC Marathon",
		ConsentGiven: true,
	}
	myConsentBytes, _ := json.Marshal(myConsent)

	// Save myConsent
	stub.MockTransactionStart("t1")
	retBytes, err := PutConsent(stub, caller, []string{string(myConsentBytes)})
	stub.MockTransactionEnd("t1")
	common_test_utils.AssertTrue(t, len(retBytes) == 0, "Expected no bytes returned")
	common_test_utils.AssertTrue(t, err == nil, "Expected no error returned")
}

