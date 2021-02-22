package gatewayapi

import (
	"errors"
	"math/big"
	"net"
	"strconv"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

func handleSingleCIDOffersPublishRequest(conn net.Conn, request *fcrmessages.FCRMessage) error {
	// Get tthe core structure
	c := core.GetSingleInstance()

	cidMin, cidMax, registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof, err := fcrmessages.DecodeGatewaySingleCIDOfferPublishRequest(request)
	if err != nil {
		return err
	}
	// TODO: Need to check registration info somewhere
	log.Info("Registration info: %v, %v, %v, %v", registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof)

	// Search offers, TODO: Use a more efficient way, maybe use separate class for storing dht request
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

	// TODO: This is not an efficient way of doing single cid offer search.
	// A better way would be storing the single cid offers separately (in a binary tree for example)
	maxMsg := 10
	maxOffer := 50
	offers := make([]fcrmessages.FCRMessage, 0)
	var offer []cidoffer.CidGroupOffer
	index1 := 0
	index2 := 0

	for i := min; i <= max; i++ {
		cid, err := cid.NewContentID(big.NewInt(i))
		if err != nil {
			return err
		}
		cidOffers, exists := c.Offers.GetOffers(cid)
		if !exists {
			continue
		}
		// Check offer one by one
		for _, cidOffer := range cidOffers {
			if len(cidOffer.Cids) == 1 {
				// It's a single cid offer
				if index2 == 0 {
					offer = make([]cidoffer.CidGroupOffer, 0)
				}
				offer = append(offer, *cidOffer)
				index2++
				if index2 >= maxOffer {
					// Encode msg
					msg, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offer)
					if err != nil {
						return err
					}
					offers = append(offers, *msg)
					index2 = 0
					index1++
					if index1 >= maxMsg {
						break
					}
				}
			}
		}
	}
	if index2 != 0 {
		// Encode msg
		msg, err := fcrmessages.EncodeProviderDHTPublishGroupCIDRequest(1, c.ProviderID, offer)
		if err != nil {
			return err
		}
		offers = append(offers, *msg)
	}

	// Send response
	response, err := fcrmessages.EncodeGatewaySingleCIDOfferPublishResponse(offers)
	if err != nil {
		return err
	}
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

	// TODO: Verifying & storing acknowledgements, will need gatewayID
	return nil
}
