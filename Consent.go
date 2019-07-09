package consent

import (
	"common/bchcls/asset_mgmt"
	"common/bchcls/custom_errors"
	"common/bchcls/global_data"
	"common/bchcls/index"
	"common/bchcls/user_mgmt"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/pkg/errors"
	"strconv"
)

var logger = shim.NewLogger("Consent")

const consentAssetNamespace string = "consent.Consent"
const IndexConsent = "Consent"

type Consent struct {
	ConsentID    string `json:"Consent_ID"`
	Name         string `json:"Name"`
	Email        string `json:"Email"`
	Phone        int    `json:"phone"`
	EventName    string `json:"EventName"`
	ConsentGiven bool   `json:"ConsentGiven"`
}

// SetupIndex creates indices for the Consent package
func SetupIndex(stub shim.ChaincodeStubInterface) error {

	consentTable := index.GetTable(stub, IndexConsent, "Consent_ID")
	err := consentTable.AddIndex([]string{"Name", "EventName", consentTable.GetPrimaryKeyId()}, false)
	if err != nil {
		err = errors.Wrap(err, "Failed to AddIndex Name,EventName to IndexConsent")
		logger.Error(err)
		return err
	}
	err = consentTable.SaveToLedger()
	if err != nil {
		err = errors.Wrap(err, "Failed to SaveToLedger for IndexConsent")
		logger.Error(err)
		return err
	}
	return nil
}

// PutConsent adds or updates a Consent on the ledger
func PutConsent(stub shim.ChaincodeStubInterface, caller global_data.User, args []string) ([]byte, error) {

	// Extract vehicle from args
	if len(args) != 1 {
		err := &custom_errors.LengthCheckingError{Type: "PutConsent arguments length"}
		logger.Error(err)
		return nil, errors.WithStack(err)
	}
	consent := Consent{}
	err := json.Unmarshal([]byte(args[0]), &consent)
	if err != nil {
		err = errors.Wrap(err, "Failed to unmarshal Consent")
		logger.Error(err)
		return nil, err
	}

	// Convert vehicle to asset
	asset := convertToAsset(consent)
	assetManager := asset_mgmt.GetAssetManager(stub, caller)

	// Encrypt the asset w/ the caller's sym key
	assetKey := user_mgmt.GetSymKey(caller)
	return nil, assetManager.AddAsset(asset, assetKey, false)
}

// GetConsentPage gets a page of Consents (that the caller has access to) from the ledger
// Inputs:
//      args: [limit, previousKey]
//      limit: the number of consents per page
//      previousKey: the key returned by the previous call to GetConsentPage
// Return values:
//      byte slice: map{vehiclePage, previousKey}. Pass the previousKey the next time you call this function.
//      error: any error that occurs
func GetConsentPage(stub shim.ChaincodeStubInterface, caller global_data.User, args []string) ([]byte, error) {

	// Parse the args
	if len(args) != 2 {
		logger.Error("Invalid args length")
		return nil, errors.New("Invalid args length")
	}
	limit, err := strconv.Atoi(args[0])
	if err != nil {
		err = errors.Wrap(err, "Invalid limit")
		logger.Error(err)
		return nil, err
	}
	previousKey := args[1]

	// Get the vehicles page
	consentPage, newPreviousKey, err := GetConsentPageInternal(stub, caller, limit, previousKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// Marshal vehicle page & previousKey into []byte
	retMap := map[string]interface{}{}
	retMap["consentPage"] = consentPage
	retMap["lastKey"] = newPreviousKey
	return json.Marshal(retMap)
}

// GetConsentInternal gets a page of Consents (that the caller has access to) from the ledger
// Inputs:
//      limit: the number of consents per page
//      previousKey: the key returned by the previous call to GetconsentsPage
// Return values:
//      Consent slice: page of Consents
//      previousKey: pass this the next time you call this function
//      error: any error that occurs
func GetConsentPageInternal(stub shim.ChaincodeStubInterface, caller global_data.User, limit int, previousKey string) ([]Consent, string, error) {

	// Get a page of vehicle assets
	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	assetPage, newPreviousKey, err := assetManager.GetAssetPage(
		consentAssetNamespace,
		IndexConsent,
		[]string{},
		[]string{},
		[]string{},
		true,
		previousKey,
		limit,
		nil)
	if err != nil {
		err = errors.Wrap(err, "Failed to GetAssetPage")
		logger.Error(err)
		return []Consent{}, "", err
	}

	// Convert from asset to Vehicle
	consentPage := []Consent{}
	for _, asset := range assetPage {
		consentPage = append(consentPage, convertFromAsset(&asset))
	}

	return consentPage, newPreviousKey, nil
}

// private function that converts vehicle to asset
func convertToAsset(consent Consent) global_data.Asset {
	asset := global_data.Asset{}
	asset.AssetId = asset_mgmt.GetAssetId(consentAssetNamespace, consent.ConsentID)
	asset.Datatypes = []string{}
	var publicData interface{}
	asset.PublicData, _ = json.Marshal(&publicData)
	asset.PrivateData, _ = json.Marshal(&consent)
	asset.IndexTableName = IndexConsent
	return asset
}

// private function that converts asset to vehicle
func convertFromAsset(asset *global_data.Asset) Consent {
	consent := Consent{}
	json.Unmarshal(asset.PrivateData, &consent)
	return consent
}
