package adminapi

import (
	"bytes"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleProviderPublishGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage, c *core.Core) {
	if c.ProviderPrivateKey == nil {
		logging.Error("This provider hasn't been initialised by the admin.")
		return
	}
	logging.Info("handleProviderPublishGroupCID: %+v", request)

	cids, price, expiry, qos, err := fcrmessages.DecodeProviderAdminPublishGroupCIDRequest(request)
	if err != nil {
		logging.Error("Error in decoding the incoming request")
		return
	}
	offer, err := cidoffer.NewCidGroupOffer(c.ProviderID, &cids, price, expiry, qos)
	if err != nil {
		logging.Error("Error in creating offer")
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
	digest := offer.GetMessageDigest()

	// TODO Add nonce.
	offerMsg, err := fcrmessages.EncodeProviderPublishGroupCIDRequest(1, offer)
	if err != nil {
		logging.Error("Error in encoding msg")
		return
	}

	// Add offer to storage
	c.GroupOffers.Add(offer)

	c.RegisteredGatewaysMapLock.RLock()
	defer c.RegisteredGatewaysMapLock.RUnlock()

	for _, gw := range c.RegisteredGatewaysMap {
		gatewayID, err := nodeid.NewNodeIDFromString(gw.GetNodeID())
		if err != nil {
			logging.Error("Error with nodeID %v: %v", gw.GetNodeID(), err)
			continue
		}

		gComm, err := c.GatewayCommPool.GetConnForRequestingNode(gatewayID, fcrtcpcomms.AccessFromProvider)
		if err != nil {
			logging.Error("Error with getting gComm with nodeID %v: %v", gw.GetNodeID(), err)
			continue
		}
		gComm.CommsLock.Lock()
		defer gComm.CommsLock.Unlock()

		err = fcrtcpcomms.SendTCPMessage(gComm.Conn, offerMsg, settings.DefaultTCPInactivityTimeout)
		if err != nil {
			logging.Error("Error with send message: %v", err)
			continue
		}

		response, err := fcrtcpcomms.ReadTCPMessage(gComm.Conn, settings.DefaultTCPInactivityTimeout)
		if err != nil {
			logging.Error("Error with receiving message: %v", err)
		}

		logging.Info("Got reponse from gateway=%v: %+v", gatewayID.ToString(), response)
		_, candidate, err := fcrmessages.DecodeProviderPublishGroupCIDResponse(response)
		if err != nil {
			logging.Error("Error with decode response: %v", err)
			continue
		}
		logging.Info("Received digest: %v", candidate)

		if bytes.Equal(candidate[:], digest[:]) {
			logging.Info("Digest is OK! Add offer to storage")
			c.NodeOfferMapLock.Lock()
			defer c.NodeOfferMapLock.Unlock()
			sentOffers, ok := c.NodeOfferMap[gatewayID.ToString()]
			if !ok {
				sentOffers = make([]cidoffer.CidGroupOffer, 0)
			}
			sentOffers = append(sentOffers, *offer)
			c.NodeOfferMap[gatewayID.ToString()] = sentOffers
		} else {
			logging.Info("Digest is not OK")
		}
	}

	response, err := fcrmessages.EncodeProviderAdminPublishOfferAck(true)
	if err != nil {
		logging.Error("Error in encoding response.")
		return
	}

	w.WriteJson(response)
}
