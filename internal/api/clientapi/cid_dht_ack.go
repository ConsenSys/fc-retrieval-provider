package clientapi

import (
	"net/http"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	log "github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleClientCIDGroupPublishDHTAckRequest(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()

	cid, gatewayID, err := fcrmessages.DecodeClientCIDGroupPublishDHTAckRequest(request)
	if err != nil {
		s := "Client DHT Ack Request: Failed to decode payload."
		log.Error(s + err.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	c.AcknowledgementMapLock.RLock()
	defer c.AcknowledgementMapLock.RUnlock()

	// Construct message
	dummyMsg, _ := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(0, c.ProviderID, make([]cidoffer.CidGroupOffer, 0))
	dummyAck, _ := fcrmessages.EncodeProviderDHTPublishGroupCIDAck(0, "")
	response, _ := fcrmessages.EncodeClientCIDGroupPublishDHTAckResponse(cid, gatewayID, false, dummyMsg, dummyAck)

	gateways, ok := c.AcknowledgementMap[cid.ToString()]
	if ok {
		ack, ok := gateways[gatewayID.ToString()]
		if ok {
			// Found an ack, update response
			response, err = fcrmessages.EncodeClientCIDGroupPublishDHTAckResponse(cid, gatewayID, true, &ack.Msg, &ack.MsgAck)
			if err != nil {
				s := "Internal error: Error encoding response."
				log.Error(s + err.Error())
				rest.Error(w, s, http.StatusBadRequest)
				return
			}
		}
	}

	// Respond
	w.WriteJson(response)
}
