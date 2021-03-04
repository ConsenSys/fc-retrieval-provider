package adminapi

import (
	"math/rand"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/providerapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleProviderDHTPublishGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()
	if c.ProviderPrivateKey == nil {
		logging.Error("This provider hasn't been initialised by the admin.")
		return
	}
	logging.Info("handleProviderDHTPublishGroupCID : %+v", request)

	cids, price, expiry, qos, err := fcrmessages.DecodeProviderAdminDHTPublishGroupCIDRequest(request)
	if err != nil {
		logging.Error("Error in decoding the incoming request ", err.Error())
		return
	}
	if len(cids) == 0 || len(cids) != len(price) || len(cids) != len(expiry) || len(cids) != len(qos) {
		logging.Error("Incoming offer info does not have same length/have zero length")
		return
	}

	offers := make([]cidoffer.CidGroupOffer, 0)
	for i := 0; i < len(cids); i++ {
		offer, err := cidoffer.NewCidGroupOffer(c.ProviderID, &[]cid.ContentID{cids[i]}, price[i], expiry[i], qos[i])
		if err != nil {
			logging.Error("Error in creating new offer ", err.Error())
			return
		}
		// Sign the offer
		err = offer.SignOffer(func(msg interface{}) (string, error) {
			return fcrcrypto.SignMessage(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion, msg)
		})
		if err != nil {
			logging.Error("Error in signing the offer.")
			return
		}
		// Append offer
		offers = append(offers, *offer)
	}

	// Add offers
	for _, offer := range offers {
		c.SingleOffers.Add(&offer)
	}

	c.RegisteredGatewaysMapLock.RLock()
	defer c.RegisteredGatewaysMapLock.RUnlock()

	for _, gw := range c.RegisteredGatewaysMap {
		gatewayID, err := nodeid.NewNodeIDFromString(gw.GetNodeID())
		if err != nil {
			logging.Error("Error with nodeID %v: %v", gw.GetNodeID(), err)
			continue
		}
		// TODO, Need to select only cid offers that are close to the gatewayID
		// For now, it selects a random offer from the offers.
		offer := offers[rand.Intn(len(offers))]
		err = providerapi.RequestDHTProviderPublishGroupCID([]cidoffer.CidGroupOffer{offer}, gatewayID)
		if err != nil {
			logging.Error("Error in publishing group offer :%v", err)
		}
	}

	// Respond to admin
	response, err := fcrmessages.EncodeProviderAdminPublishOfferAck(true)
	if err != nil {
		logging.Error("Error in encoding response.")
		return
	}
	// Sign the response
	response.SignMessage(func(msg interface{}) (string, error) {
		return fcrcrypto.SignMessage(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion, msg)
	})
	w.WriteJson(response)
}