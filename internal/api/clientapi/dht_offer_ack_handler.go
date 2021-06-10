/*
Package clientapi - set of remote API used to call a Retrieval Provider, grouped to a specific caller type - Retrieval Client.
All calls from FileCoin Secondary Retrieval network nodes of type Retrieval Client are going to API handlers in this package.
*/
package clientapi

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

// HandleClientDHTOfferAckRequest is used to handle client request for ack a dht offer
func HandleClientDHTOfferAckRequest(w rest.ResponseWriter, request *fcrmessages.FCRMessage) {
	// Get core structure
	c := core.GetSingleInstance()

	cid, gatewayID, err := fcrmessages.DecodeClientDHTOfferAckRequest(request)
	if err != nil {
		s := "Fail to decode message."
		logging.Error(s + err.Error())
		rest.Error(w, s, http.StatusBadRequest)
		return
	}

	c.AcknowledgementMapLock.RLock()
	defer c.AcknowledgementMapLock.RUnlock()

	// Construct message
	dummyMsg, _ := fcrmessages.EncodeProviderPublishDHTOfferRequest(c.ProviderID, 0, make([]cidoffer.CIDOffer, 0))
	dummyAck, _ := fcrmessages.EncodeProviderPublishDHTOfferResponse(0, "")
	response, _ := fcrmessages.EncodeClientDHTOfferAckResponse(cid, gatewayID, false, dummyMsg, dummyAck)

	gateways, ok := c.AcknowledgementMap[cid.ToString()]
	if ok {
		ack, ok := gateways[gatewayID.ToString()]
		if ok {
			// Found an ack, update response
			response, err = fcrmessages.EncodeClientDHTOfferAckResponse(cid, gatewayID, true, &ack.Msg, &ack.MsgAck)
			if err != nil {
				s := "Internal error: Fail to encode message."
				logging.Error(s + err.Error())
				rest.Error(w, s, http.StatusInternalServerError)
				return
			}
		}
	}

	// Sign message
	if signErr := response.Sign(c.ProviderPrivateKey, c.ProviderPrivateKeyVersion); signErr != nil {
		s := "Internal error: Fail to sign message."
		logging.Error(s + signErr.Error())
		rest.Error(w, s, http.StatusInternalServerError)
		return
	}

  if writeErr := w.WriteJson(response); writeErr != nil {
    logging.Error("can't write JSON during HandleClientDHTOfferAckRequest %s", writeErr.Error())
  }
}
