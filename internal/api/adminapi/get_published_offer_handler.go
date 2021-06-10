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
  "github.com/ConsenSys/fc-retrieval-provider/internal/core"
)

// HandleProviderAdminGetPublishedOfferRequest handles provider admin publish offer request
func HandleProviderAdminGetPublishedOfferRequest(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()

	if c.ProviderPrivateKey == nil {
		s := "This provider hasn't been initialised by the admin"
		logging.Error(s)
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	gatewayIDs, err := fcrmessages.DecodeProviderAdminGetPublishedOfferRequest(request)
	if err != nil {
		s := "Fail to decode message."
		logging.Error(s + err.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	offers := make([]cidoffer.CIDOffer, 0)
	c.NodeOfferMapLock.Lock()
	defer c.NodeOfferMapLock.Unlock()
	if len(gatewayIDs) > 0 {
		for _, gatewayID := range gatewayIDs {
			offs := c.NodeOfferMap[gatewayID.ToString()]
			for _, off := range offs {
				offers = append(offers, off)
			}
		}
	} else {
		for _, values := range c.NodeOfferMap {
			for _, value := range values {
				offers = append(offers, value)
			}
		}
	}

	// Construct message
	response, err := fcrmessages.EncodeProviderAdminGetPublishedOfferResponse(
		len(offers) > 0,
		offers,
	)
	if err != nil {
		s := "Internal error: Fail to encode message."
		logging.Error(s + err.Error())
		rest.Error(w, s, http.StatusInternalServerError)
		return
	}

	// Sign message
	if signErr := response.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion); signErr != nil {
		s := "Internal error: Fail to sign message."
		logging.Error(s + signErr.Error())
		rest.Error(w, s, http.StatusInternalServerError)
		return
	}
	// Send message
  if writeErr := w.WriteJson(response); writeErr != nil {
    logging.Error("can't write JSON during HandleProviderAdminGetPublishedOfferRequest %s", writeErr.Error())
  }
}
