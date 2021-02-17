package adminapi

import (
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
		err = p.SendMessageToGateway(request, gatewayID)
		if err != nil {
			logging.Error("Error with send message: %v", err)
			continue
		}
		_, offer, _ := fcrmessages.DecodeProviderPublishGroupCIDRequest(request)
		p.AppendOffer(gatewayID, offer)
	}
	w.WriteHeader(http.StatusOK)
}