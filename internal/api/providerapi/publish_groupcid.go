package providerapi

import (
	"bytes"
	"errors"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
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

	// Get the gateways's signing key
	c.RegisteredGatewaysMapLock.RLock()
	defer c.RegisteredGatewaysMapLock.RUnlock()
	pubKey, err := c.RegisteredGatewaysMap[gatewayID.ToString()].GetSigningKey()
	if err != nil {
		return err
	}

	// Construct message, TODO: Add nonce
	request, err := fcrmessages.EncodeProviderPublishGroupCIDRequest(1, offer)
	if err != nil {
		return err
	}
	// Sign the request
	request.SignMessage(func(msg interface{}) (string, error) {
		return fcrcrypto.SignMessage(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion, msg)
	})
	// Send request
	err = fcrtcpcomms.SendTCPMessage(gComm.Conn, request, settings.DefaultTCPInactivityTimeout)
	if err != nil {
		c.GatewayCommPool.DeregisterNodeCommunication(gatewayID)
		return err
	}
	// Get a response
	response, err := fcrtcpcomms.ReadTCPMessage(gComm.Conn, settings.DefaultTCPInactivityTimeout)
	if err != nil {
		c.GatewayCommPool.DeregisterNodeCommunication(gatewayID)
		return err
	}
	// Verify the response
	ok, err := response.VerifySignature(func(sig string, msg interface{}) (bool, error) {
		return fcrcrypto.VerifyMessage(pubKey, sig, msg)
	})
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Fail to verify the response")
	}
	logging.Info("Got reponse from gateway=%v: %+v", gatewayID.ToString(), response)
	// TODO: Check nonce
	_, candidate, err := fcrmessages.DecodeProviderPublishGroupCIDResponse(response)
	if err != nil {
		logging.Error("Error with decode response: %v", err)
		return err
	}
	logging.Info("Received digest: %v", candidate)

	digest := offer.GetMessageDigest()

	// Add offer to storage
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
		return errors.New("Digest not match")
	}
	return nil
}
