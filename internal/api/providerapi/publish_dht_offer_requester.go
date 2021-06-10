/*
Package providerapi - set of remote API used to call a Retrieval Provider, grouped to a specific caller type - Retrieval Provider.
All calls from FileCoin Secondary Retrieval network nodes of type Retrieval Provider are going to API handlers in this package.
*/
package providerapi

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

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrp2pserver"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
)

// RequestProviderPublishDHTOffer is used to publish a dht offer to a given gateway
func RequestProviderPublishDHTOffer(reader *fcrp2pserver.FCRServerReader, writer *fcrp2pserver.FCRServerWriter, args ...interface{}) (*fcrmessages.FCRMessage, error) {
	// Get parameters
	if len(args) != 2 {
		return nil, errors.New("wrong arguments")
	}
	offers, ok := args[0].([]cidoffer.CIDOffer)
	if !ok {
		return nil, errors.New("wrong arguments")
	}
	gatewayID, ok := args[1].(*nodeid.NodeID)
	if !ok {
		return nil, errors.New("wrong arguments")
	}

	// Get the core structure
	c := core.GetSingleInstance()

	// Construct message
	// TODO: Add nonce
	request, err := fcrmessages.EncodeProviderPublishDHTOfferRequest(c.ProviderID, 1, offers)
	if err != nil {
		return nil, err
	}
	// Sign the request
	if request.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion) != nil {
		return nil, errors.New("internal error in signing the message")
	}
	// Send the request
	err = writer.Write(request, c.Settings.TCPInactivityTimeout)
	if err != nil {
		return nil, err
	}
	// Get a response
	response, err := reader.Read(c.Settings.TCPInactivityTimeout)
	if err != nil {
		return nil, err
	}

	// Verify the response
	// Get the gateway's signing key
	gatewayInfo := c.RegisterMgr.GetGateway(gatewayID)
	if gatewayInfo == nil {
		return nil, errors.New("gateway information not found")
	}
	pubKey, err := gatewayInfo.GetSigningKey()
	if err != nil {
		return nil, errors.New("fail to obatin the public key")
	}

	if response.Verify(pubKey) != nil {
		return nil, errors.New("fail to verify the response")
	}

	// Verify the acks
	// TODO: Check nonce
	_, sig, err := fcrmessages.DecodeProviderPublishDHTOfferResponse(response)
	if err != nil {
		return nil, err
	}
	ok, err = fcrcrypto.VerifyMessage(pubKey, sig, request)
	if err != nil {
		return nil, errors.New("internal error in verifying ack")
	}
	if !ok {
		return nil, errors.New("fail to verify the ack")
	}

	// Add offer to ack map and storage
	for _, offer := range offers {
		// Add offer to storage
		c.NodeOfferMapLock.Lock()
		sentOffers, ok := c.NodeOfferMap[gatewayID.ToString()]
		if !ok {
			sentOffers = make([]cidoffer.CIDOffer, 0)
		}
		sentOffers = append(sentOffers, offer)
		c.NodeOfferMap[gatewayID.ToString()] = sentOffers
		c.NodeOfferMapLock.Unlock()
		// Add offer to ack map
		c.AcknowledgementMapLock.Lock()
		cidMap, ok := c.AcknowledgementMap[(offer.GetCIDs())[0].ToString()]
		if !ok {
			cidMap = make(map[string]core.DHTAcknowledgement)
			c.AcknowledgementMap[(offer.GetCIDs())[0].ToString()] = cidMap
		}
		cidMap[gatewayID.ToString()] = core.DHTAcknowledgement{
			Msg:    *request,
			MsgAck: *response,
		}
		c.AcknowledgementMapLock.Unlock()
	}
	return response, nil
}
