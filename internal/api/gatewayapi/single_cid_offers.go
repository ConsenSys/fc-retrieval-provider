package gatewayapi

import (
	"errors"
	"math/big"
	"net"
	"strconv"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

func handleSingleCIDOffersPublishRequest(conn net.Conn, request *fcrmessages.FCRMessage) error {
	// Get the core structure
	c := core.GetSingleInstance()
	gatewayID, cidMin, cidMax, registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof, err := fcrmessages.DecodeGatewaySingleCIDOfferPublishRequest(request)
	if err != nil {
		return err
	}
	// TODO, Need to check registration info
	log.Info("Registration info: %v, %v, %v, %v", registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof)
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
			msg, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offers[:50]) //TODO, nonce?
			if err != nil {
				return err
			}
			msgs = append(msgs, *msg)
			offers = offers[50:]
		} else {
			msg, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offers) //TODO, nonce?
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
	acknowledgements, err := fcrmessages.DecodeGatewaySingleCIDOfferPublishResponseAck(acks)
	if len(acknowledgements) != len(offers) {
		return errors.New("Invalid response")
	}
	for i, acknowledgement := range acknowledgements {
		nonce, signature, err := fcrmessages.DecodeProviderDHTPublishGroupCIDAck(&acknowledgement)
		if err != nil {
			return err
		}
		if nonce != 1 { // TODO, add nonce
			return errors.New("Nonce mismatch")
		}
		c.RegisteredGatewaysMapLock.RLock()
		key, err := c.RegisteredGatewaysMap[gatewayID.ToString()].GetSigningKey()
		if err != nil {
			return err
		}
		ok, err := fcrcrypto.VerifyMessage(key, signature, msgs[i])
		c.RegisteredGatewaysMapLock.RUnlock()
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
