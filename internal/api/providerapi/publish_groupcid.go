package providerapi

import (
	"bytes"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

// RequestProviderPublishGroupCID is used to publish a group CID offer to a given gateway
func RequestProviderPublishGroupCID(offer *cidoffer.CidGroupOffer, gatewayID *nodeid.NodeID) error {
	// Get the core structure
	c := core.GetSingleInstance()

	// Get the connection to the given gateway
	gComm, err := c.GatewayCommPool.GetConnForRequestingNode(gatewayID, fcrtcpcomms.AccessFromProvider)
	if err != nil {
		return err
	}
	gComm.CommsLock.Lock()
	defer gComm.CommsLock.Unlock()
	// Construct message
	request, err := fcrmessages.EncodeProviderPublishGroupCIDRequest(1, offer)
	if err != nil {
		return err
	}
	err = fcrtcpcomms.SendTCPMessage(gComm.Conn, request, settings.DefaultTCPInactivityTimeout)
	if err != nil {
		c.GatewayCommPool.DeregisterNodeCommunication(gatewayID)
		return err
	}

	// Get a response
	response, err := fcrtcpcomms.ReadTCPMessage(gComm.Conn, settings.DefaultTCPInactivityTimeout)
	if err != nil {
		return err
	}

	log.Info("Got reponse from gateway=%v: %+v", gatewayID.ToString(), response)
	_, candidate, err := fcrmessages.DecodeProviderPublishGroupCIDResponse(response)
	if err != nil {
		return err
	}
	log.Info("Received digest: %v", candidate)
	digest := offer.GetMessageDigest()
	if bytes.Equal(candidate[:], digest[:]) {
		log.Info("Digest is OK! Add offer to storage")
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
		log.Info("Digest is not OK")
	}

	return nil
}
