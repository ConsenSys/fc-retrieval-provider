/*
Package gatewayapi - set of remote API used to call a Retrieval Provider, grouped to a specific caller type - Retrieval Gateway.
All calls from FileCoin Secondary Retrieval network nodes of type Retrieval Gateway are going to API handlers in this package.
*/
package gatewayapi

/*
 * Copyright 2020 ConsenSys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"errors"

	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrp2pserver"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
)

// HandleGatewayListDHTOfferRequest handles the gateway list dht offers request
func HandleGatewayListDHTOfferRequest(reader *fcrp2pserver.FCRServerReader, writer *fcrp2pserver.FCRServerWriter, request *fcrmessages.FCRMessage) error {
	// Get core structure
	c := core.GetSingleInstance()

	gatewayID, cidMin, cidMax, registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof, err := fcrmessages.DecodeGatewayListDHTOfferRequest(request)
	if err != nil {
		// Reply with invalid message
		return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
	}

	// Get the gateway's signing key
	gatewayInfo := c.RegisterMgr.GetGateway(gatewayID)
	if gatewayInfo == nil {
		logging.Warn("Gateway information not found for %s.", gatewayID.ToString())
		return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
	}
	pubKey, err := gatewayInfo.GetSigningKey()
	if err != nil {
		logging.Warn("Fail to obtain the public key for %s", gatewayID.ToString())
		return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
	}

	// First verify the message
	if request.Verify(pubKey) != nil {
		logging.Warn("Fail to verify the request from %s", gatewayID.ToString())
		return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
	}

	// TODO, Need to check registration info
	logging.Info("Registration info: %v, %v, %v, %v", registrationBlock, registrationTransactionReceipt, registrationMerkleRoot, registrationMerkleProof)

	// Search offers
	maxOffers := 500      //TODO, max offers configurable?
	maxOffersPerMsg := 50 //TODO, max offer per message?
	msgs := make([]fcrmessages.FCRMessage, 0)
	offers, exists := c.OffersMgr.GetDHTOffersWithinRange(cidMin, cidMax, maxOffers)

	if exists {
		logging.Info("Found offers: ")
		for _, offer := range offers {
			logging.Info(offer.GetCIDs()[0].ToString())
		}
	}

	if exists {
		exit := false
		for !exit {
			var msg *fcrmessages.FCRMessage
			if len(offers) > maxOffersPerMsg {
				msg, err = fcrmessages.EncodeProviderPublishDHTOfferRequest(c.ProviderID, 1, offers[:50]) //TODO: Add nonce
				if err != nil {
					return err
				}
				offers = offers[50:]
			} else {
				msg, err = fcrmessages.EncodeProviderPublishDHTOfferRequest(c.ProviderID, 1, offers) //TODO: Add nonce
				if err != nil {
					return err
				}
				exit = true
			}
			// Sign the sub message
			if msg.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion) != nil {
				logging.Error("Internal error in signing message.")
				return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
			}
			msgs = append(msgs, *msg)
		}
	}

	// Construct response
	response, err := fcrmessages.EncodeGatewayListDHTOfferResponse(msgs)
	if err != nil {
		return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
	}
	// Sign the response
	if response.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion) != nil {
		logging.Error("Internal error in signing message.")
		return writer.WriteInvalidMessage(c.Settings.TCPInactivityTimeout)
	}
	// Respond
	err = writer.Write(response, c.Settings.TCPInactivityTimeout)
	if err != nil {
		return err
	}

	// Get acks
	acks, err := reader.Read(c.Settings.TCPLongInactivityTimeout)
	if err != nil {
		return err
	}
	// Verify the acks
	if acks.Verify(pubKey) != nil {
		return errors.New("fail to verify the acks")
	}

	acknowledgements, err := fcrmessages.DecodeGatewayListDHTOfferAck(acks)
	if len(acknowledgements) != len(msgs) {
		return errors.New("invalid response")
	}
	for i, acknowledgement := range acknowledgements {
		// TODO: Check nonce.
		_, signature, err := fcrmessages.DecodeProviderPublishDHTOfferResponse(&acknowledgement)
		if err != nil {
			return err
		}
		ok, err := fcrcrypto.VerifyMessage(pubKey, signature, msgs[i].GetMessageBody())
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("verification failed")
		}
		// It's okay, add to acknowledgements map
		_, _, sentOffers, err := fcrmessages.DecodeProviderPublishDHTOfferRequest(&msgs[i])
		if err != nil {
			return err
		}
		c.AcknowledgementMapLock.Lock()
		for _, sentOffer := range sentOffers {
			c.AcknowledgementMap[sentOffer.GetCIDs()[0].ToString()][gatewayID.ToString()] = core.DHTAcknowledgement{
				Msg:    msgs[i],
				MsgAck: acknowledgement,
			}
		}
		c.AcknowledgementMapLock.Unlock()
	}
	return nil
}
