package adminapi

import (
	"bytes"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
	"github.com/ConsenSys/fc-retrieval-provider/internal/register"
)

func handleProviderPublishGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage, p *provider.Provider) {
	logging.Info("handleProviderPublishGroupCID: %+v", request)
	gateways, err := register.GetRegisteredGateways(p)
	if err != nil {
		logging.Error("Error with get registered gateways %v", err)
		panic(err)
	}
	for _, gw := range gateways {
		gatewayID, err := nodeid.NewNodeIDFromString(gw.NodeID)
		if err != nil {
			logging.Error("Error with nodeID %v: %v", gw.NodeID, err)
			continue
		}
		response, err := p.SendMessageToGateway(request, gatewayID)
		if err != nil {
			logging.Error("Error with send message: %v", err)
			continue
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
			p.AppendOffer(gatewayID, offer)
		} else {
			logging.Info("Digest is not OK")
		}
	}
	w.WriteHeader(http.StatusOK)
}