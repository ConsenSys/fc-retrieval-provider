package adminapi

import (
	"bytes"
	"net/http"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleProviderPublishGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage, c *core.Core) {
	logging.Info("handleProviderPublishGroupCID: %+v", request)

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

		err = fcrtcpcomms.SendTCPMessage(gComm.Conn, request, settings.DefaultTCPInactivityTimeout)
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
		_, offer, _ := fcrmessages.DecodeProviderPublishGroupCIDRequest(request)
		digest := offer.GetMessageDigest()
		if bytes.Equal(candidate[:], digest[:]) {
			logging.Info("Digest is OK! Add offer to storage")
			// TODO
			c.GroupOffers.Add(offer)
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
	w.WriteHeader(http.StatusOK)
}
