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

  "github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
  "github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
  "github.com/ConsenSys/fc-retrieval-common/pkg/logging"
  "github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
  "github.com/ConsenSys/fc-retrieval-provider/internal/core"
)

// HandleProviderAdminPublishGroupOfferRequest handles provider admin publish group offer request
func HandleProviderAdminPublishGroupOfferRequest(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()

	if c.ProviderPrivateKey == nil {
		s := "This provider hasn't been initialised by the admin"
		logging.Error(s)
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	cids, price, expiry, qos, err := fcrmessages.DecodeProviderAdminPublishGroupOfferRequest(request)
	if err != nil {
		s := "Fail to decode message."
		logging.Error(s + err.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	offer, err := cidoffer.NewCIDOffer(c.ProviderID, cids, price, expiry, qos)
	if err != nil {
		s := "Fail to generate offer."
		logging.Error(s + err.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}
	// Sign the offer
	if signErr := offer.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion); signErr != nil {
		s := "Fail to sign offer."
		logging.Error(s + signErr.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	// Add offer to storage
	if addErr := c.OffersMgr.AddGroupOffer(offer); addErr != nil {
	  logging.Error("can't add group offer: %v", offer)
  }

	// Get all gateways
	gateways := c.RegisterMgr.GetAllGateways()
	for _, gateway := range gateways {
		gatewayID, err := nodeid.NewNodeIDFromHexString(gateway.GetNodeID())
		if err != nil {
			s := "Fail to generate node id."
			logging.Error(s + err.Error())
			rest.Error(w, s, http.StatusBadRequest)
			return
		}
		_, err = c.P2PServer.RequestGatewayFromProvider(gatewayID, fcrmessages.ProviderPublishGroupOfferRequestType, offer, gatewayID)
		if err != nil {
			logging.Error("Error in publishing group offer to %s: %s", gatewayID.ToString(), err.Error())
		}
	}

	// Respond to admin
	response, err := fcrmessages.EncodeProviderAdminPublishGroupOfferResponse(true)
	if err != nil {
		logging.Error("Error in encoding response.")
		return
	}
	// Sign the response
	if response.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion) != nil {
		logging.Error("Error in signing the response.")
		return
	}

  if writeErr := w.WriteJson(response); writeErr != nil {
    logging.Error("can't write JSON during HandleProviderAdminPublishGroupOfferRequest %s", writeErr.Error())
  }
}
