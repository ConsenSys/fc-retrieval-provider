package gatewayapi

import (
	"errors"
	"math/big"
	"net"
	"strconv"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

func handleSingleCIDOffersPublishRequest(conn net.Conn, request *fcrmessages.FCRMessage) error {
	// Get core structure
	c := core.GetSingleInstance()
	gatewayID, cidMin, cidMax, registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof, err := fcrmessages.DecodeGatewaySingleCIDOfferPublishRequest(request)
	if err != nil {
		return err
	}
	// TODO, Need to check registration info
	log.Info("Registration info: %v, %v, %v, %v", registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof)

	// Get the gateways's signing key
	c.RegisteredGatewaysMapLock.RLock()
	defer c.RegisteredGatewaysMapLock.RUnlock()
	_, ok := c.RegisteredGatewaysMap[gatewayID.ToString()]
	if !ok {
		return errors.New("Gateway register not found")
	}
	pubKey, err := c.RegisteredGatewaysMap[gatewayID.ToString()].GetSigningKey()
	if err != nil {
		return err
	}
	// Verify the incoming request
	ok, err = request.VerifySignature(func(sig string, msg interface{}) (bool, error) {
		return fcrcrypto.VerifyMessage(pubKey, sig, msg)
	})
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Fail to verify the request")
	}

	// Search offers
	min, err := strconv.ParseInt(cidMin.ToString(), 16, 32) // TODO, CHECK IF THIS IS CORRECT
	if err != nil {
		return err
	}
	max, err := strconv.ParseInt(cidMax.ToString(), 16, 32) // TODO, CHECK IF THIS IS CORRECT
	if err != nil {
		return err
	}
	if max < min {
		return errors.New("Invalid parameters")
	}

	maxOffers := 500
	offers := make([]cidoffer.CidGroupOffer, 0)
	for i := min; i <= max; i++ {
		id, err := cid.NewContentID(big.NewInt(i))
		if err != nil {
			return err
		}
		offers, exists := c.SingleOffers.GetOffers(id)
		if exists {
			for _, offer := range offers {
				offers = append(offers, offer)
				if len(offers) >= maxOffers {
					break
				}
			}
		}
		if len(offers) >= maxOffers {
			break
		}
	}
	maxOffersPerMsg := 50
	msgs := make([]fcrmessages.FCRMessage, 0)
	for {
		if len(offers) > maxOffersPerMsg {
			msg, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offers[:50]) //TODO: Add nonce
			if err != nil {
				return err
			}
			msgs = append(msgs, *msg)
			offers = offers[50:]
		} else {
			msg, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offers) //TODO: Add nonce
			if err != nil {
				return err
			}
			msgs = append(msgs, *msg)
			break
		}
	}

	// Construct response
	response, err := fcrmessages.EncodeGatewaySingleCIDOfferPublishResponse(msgs)
	if err != nil {
		return err
	}
	// Sign the response
	response.SignMessage(func(msg interface{}) (string, error) {
		return fcrcrypto.SignMessage(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion, msg)
	})
	// Respond
	err = fcrtcpcomms.SendTCPMessage(conn, response, settings.DefaultLongTCPInactivityTimeout)
	if err != nil {
		return err
	}

	// Get acks
	acks, err := fcrtcpcomms.ReadTCPMessage(conn, settings.DefaultLongTCPInactivityTimeout)
	if err != nil {
		return err
	}
	// Verify the acks
	ok, err = acks.VerifySignature(func(sig string, msg interface{}) (bool, error) {
		return fcrcrypto.VerifyMessage(pubKey, sig, msg)
	})
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Fail to verify the acks")
	}

	acknowledgements, err := fcrmessages.DecodeGatewaySingleCIDOfferPublishResponseAck(acks)
	if len(acknowledgements) != len(offers) {
		return errors.New("Invalid response")
	}
	for i, acknowledgement := range acknowledgements {
		// TODO: Check nonce.
		_, signature, err := fcrmessages.DecodeProviderDHTPublishGroupCIDAck(&acknowledgement)
		if err != nil {
			return err
		}
		ok, err := fcrcrypto.VerifyMessage(pubKey, signature, msgs[i])
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("Verification failed")
		}
		// It's okay, add to acknowledgements map
		c.AcknowledgementMapLock.Lock()
		c.AcknowledgementMap[offers[i].Cids[0].ToString()][gatewayID.ToString()] = core.DHTAcknowledgement{
			Msg:    msgs[i],
			MsgAck: acknowledgement,
		}
		c.AcknowledgementMapLock.Unlock()
	}
	return nil
}
