package adminapi

import (
	"net/http"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/providerapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleProviderPublishGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage, c *core.Core) {
	logging.Info("handleProviderPublishGroupCID: %+v", request)
	_, dummyOffer, _ := fcrmessages.DecodeProviderPublishGroupCIDRequest(request)

	// Create & Sign the offer
	offer, err := cidoffer.NewCidGroupOffer(dummyOffer.NodeID, &dummyOffer.Cids, dummyOffer.Price, dummyOffer.Expiry, dummyOffer.QoS)
	if err != nil {
		logging.Error("Error in creating offer.")
		return
	}
	err = offer.SignOffer(func(msg interface{}) (string, error) {
		return fcrcrypto.SignMessage(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion, msg)
	})
	if err != nil {
		logging.Error("Error in signing the offer.")
		return
	}

	c.RegisteredGatewaysMapLock.RLock()
	defer c.RegisteredGatewaysMapLock.RUnlock()

	for _, gw := range c.RegisteredGatewaysMap {
		gatewayID, err := nodeid.NewNodeIDFromString(gw.GetNodeID())
		if err != nil {
			logging.Error("Error with nodeID %v: %v", gw.GetNodeID(), err)
			continue
		}
		err = providerapi.RequestProviderPublishGroupCID(offer, gatewayID)
		if err != nil {
			logging.Error("Error in sending group cid offer: %v", err)
			continue
		}
	}
	w.WriteHeader(http.StatusOK)
}
