package adminapi

import (
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleKeyManagement(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()
	logging.Info("handle key management.")

	nodeID, encprivatekey, encprivatekeyversion, err := fcrmessages.DecodeAdminAcceptKeyChallenge(request)
	if err != nil {
		logging.Error("Error in decoding message.")
		return
	}

	// Decode private key from hex string to *fcrCrypto.KeyPair
	privatekey, err := fcrcrypto.DecodePrivateKey(encprivatekey)
	if err != nil {
		logging.Error("Error in decoding private key")
		return
	}

	// Decode from int32 to *fcrCrypto.KeyVersion
	privatekeyversion := fcrcrypto.DecodeKeyVersion(encprivatekeyversion)

	// Set the node id
	logging.Info("Check if c is nil :%v", c == nil)
	logging.Info("Setting node id")
	c.ProviderID = nodeID
	c.ProviderPrivateKey = privatekey
	c.ProviderPrivateKeyVersion = privatekeyversion

	// Construct messaqe
	exists := true
	response, err := fcrmessages.EncodeAdminAcceptKeyResponse(exists)
	if err != nil {
		logging.Error("Error in encoding message")
		return
	}

	logging.Info("Signing response.")
	// Sign the response
	response.SignMessage(func(msg interface{}) (string, error) {
		return fcrcrypto.SignMessage(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion, msg)
	})
	w.WriteJson(response)
}
