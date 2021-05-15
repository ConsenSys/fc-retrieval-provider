package adminapi

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
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
)

// HandleProviderAdminPublishDHTOfferRequest handles provider admin publish dht offer request
func HandleProviderAdminPublishDHTOfferRequest(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()

	if c.ProviderPrivateKey == nil {
		s := "This provider hasn't been initialised by the admin"
		logging.Error(s)
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	cids, price, expiry, qos, err := fcrmessages.DecodeProviderAdminPublishDHTOfferRequest(request)
	if err != nil {
		s := "Fail to decode message."
		logging.Error(s + err.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	if len(cids) == 0 || len(cids) != len(price) || len(cids) != len(expiry) || len(cids) != len(qos) {
		s := "Incoming offer info does not have same length/have zero length"
		logging.Error(s)
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	offers := make([]cidoffer.CIDOffer, 0)
	for i := 0; i < len(cids); i++ {
		offer, err := cidoffer.NewCIDOffer(c.ProviderID, []cid.ContentID{cids[i]}, price[i], expiry[i], qos[i])
		if err != nil {
			s := "Internal error: Fail to create new offer."
			logging.Error(s + err.Error())
			rest.Error(w, s, http.StatusInternalServerError)
			return
		}
		// Sign the offer
		if offer.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion) != nil {
			s := "Internal error: Fail to sign offer."
			logging.Error(s + err.Error())
			rest.Error(w, s, http.StatusInternalServerError)
			return
		}
		// Append offer
		offers = append(offers, *offer)
	}

	// Add offers
	for _, offer := range offers {
		c.OffersMgr.AddDHTOffer(&offer)
	}

	for _, cid := range cids {

		//TODO: merge - remove nil;  (*cid.ContentID, int, []*nodeid.NodeID)
		gateways, err := c.RegisterMgr.GetGatewaysNearCID(&cid, 0, nil)

		if err != nil {
			s := "Internal error: Fail to get gateways near the given cid."
			logging.Error(s + err.Error())
			rest.Error(w, s, http.StatusInternalServerError)
			return
		}
		for _, gw := range gateways {
			gatewayID, err := nodeid.NewNodeIDFromHexString(gw.NodeID)
			if err != nil {
				s := "Fail to generate node id."
				logging.Error(s + err.Error())
				rest.Error(w, s, http.StatusBadRequest)
				return
			}
			_, err = c.P2PServer.RequestGatewayFromProvider(gatewayID, fcrmessages.ProviderPublishDHTOfferRequestType, offers, gatewayID)
			if err != nil {
				logging.Error("Error in publishing dht offer to %s: %s", gatewayID.ToString(), err.Error())
			}
		}
	}

	// Respond to admin
	response, err := fcrmessages.EncodeProviderAdminPublishDHTOfferResponse(true)
	if err != nil {
		logging.Error("Error in encoding response.")
		return
	}
	// Sign the response
	if response.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion) != nil {
		logging.Error("Error in signing message.")
		return
	}
	w.WriteJson(response)
}
