package providerapi

import (
	"errors"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

// RequestDHTProviderPublishGroupCID is used to publish a dht group CID offer to a given gateway
func RequestDHTProviderPublishGroupCID(offers []cidoffer.CidGroupOffer, gatewayID *nodeid.NodeID) error {
	// Get the core structure
	c := core.GetSingleInstance()

	// Get the connection to the given gateway
	gComm, err := c.GatewayCommPool.GetConnForRequestingNode(gatewayID, fcrtcpcomms.AccessFromProvider)
	if err != nil {
		return err
	}

	// Construct message, TODO: Add nonce
	request, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offers)
	if err != nil {
		return err
	}
	err = fcrtcpcomms.SendTCPMessage(gComm.Conn, request, settings.DefaultTCPInactivityTimeout)
	if err != nil {
		c.GatewayCommPool.DeregisterNodeCommunication(gatewayID)
		return err
	}
	// Get a response
	response, err := fcrtcpcomms.ReadTCPMessage(gComm.Conn, settings.DefaultLongTCPInactivityTimeout)
	if err != nil {
		c.GatewayCommPool.DeregisterNodeCommunication(gatewayID)
		return err
	}
	// Need to verify the acks
	nonce, sig, err := fcrmessages.DecodeProviderDHTPublishGroupCIDAck(response)
	if err != nil {
		return err
	}

	// Check nonce
	if nonce != 1 {
		return errors.New("Nonce mismatch")
	}

	// Check signature
	// Get the public key
	c.RegisteredGatewaysMapLock.RLock()
	defer c.RegisteredGatewaysMapLock.RUnlock()
	gateway, ok := c.RegisteredGatewaysMap[gatewayID.ToString()]
	if !ok {
		return errors.New("Gateway public key not found")
	}
	pubKey, err := gateway.GetSigningKey()
	if err != nil {
		return errors.New("Fail to get signing key from gateway registration info")
	}

	ok, err = fcrcrypto.VerifyMessage(pubKey, sig, request)
	if err != nil {
		return errors.New("Internal error in verifying ack")
	}

	if !ok {
		return errors.New("Fail to verify the ack")
	}

	for _, offer := range offers {
		c.AcknowledgementMapLock.Lock()
		defer c.AcknowledgementMapLock.Unlock()
		cidMap, ok := c.AcknowledgementMap[(*offer.GetCIDs())[0].ToString()]
		if !ok {
			cidMap = make(map[string]core.DHTAcknowledgement)
			c.AcknowledgementMap[(*offer.GetCIDs())[0].ToString()] = cidMap
		}
		cidMap[gatewayID.ToString()] = core.DHTAcknowledgement{
			Msg:    *request,
			MsgAck: *response,
		}
	}
	return nil
}
