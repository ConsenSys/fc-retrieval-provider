package core

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
	"sync"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcroffermgr"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrp2pserver"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrregistermgr"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrrestserver"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"

	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

const (
	protocolVersion   = 1 // Main protocol version
	protocolSupported = 1 // Alternative protocol version
)

// DHTAcknowledgement stores the acknowledgement of a single cid offer
type DHTAcknowledgement struct {
	Msg    fcrmessages.FCRMessage // Original message
	MsgAck fcrmessages.FCRMessage // Original message ACK
}

// Core holds the main data structure for the whole provider
type Core struct {
	ProtocolVersion   int32
	ProtocolSupported []int32

	// Settings
	Settings *settings.AppSettings

	// ProviderID of this provider
	ProviderID *nodeid.NodeID

	// Provider Private Key of this provider
	ProviderPrivateKey *fcrcrypto.KeyPair

	// ProviderPrivateKeyVersion is the key version number of the private key.
	ProviderPrivateKeyVersion *fcrcrypto.KeyVersion

	// RegisterMgr manages all register related activities
	RegisterMgr *fcrregistermgr.FCRRegisterMgr

	// P2PServer handles all communication to/from gateways/providers
	P2PServer *fcrp2pserver.FCRP2PServer

	// RESTServer handles all communication to/from client/admin
	RESTServer *fcrrestserver.FCRRESTServer

	// Offer Manager
	OffersMgr *fcroffermgr.FCROfferMgr

	// Node to offer map, TODO: Use a manager
	NodeOfferMap     map[string]([]cidoffer.CIDOffer)
	NodeOfferMapLock sync.Mutex

	// Acknowledgement for every single cid offer sent (map from cid id -> map of gateway -> ack)
	AcknowledgementMap     map[string](map[string]DHTAcknowledgement)
	AcknowledgementMapLock sync.RWMutex

	// List of Gateways that allow group CID offer to be published
	GroupOfferGatewayIDs []nodeid.NodeID
}

// Single instance of the provider
var instance *Core
var doOnce sync.Once

// GetSingleInstance returns the single instance of the provider
func GetSingleInstance(confs ...*settings.AppSettings) *Core {
	doOnce.Do(func() {
		if len(confs) == 0 {
			logging.ErrorAndPanic("No settings supplied to Gateway start-up")
		}
		if len(confs) != 1 {
			logging.ErrorAndPanic("More than one sets of settings supplied to Gateway start-up")
		}

		instance = &Core{
			ProtocolVersion:           protocolVersion,
			ProtocolSupported:         []int32{protocolVersion, protocolSupported},
			Settings:                  confs[0],
			ProviderID:                nil,
			ProviderPrivateKey:        nil,
			ProviderPrivateKeyVersion: nil,
			OffersMgr:                 fcroffermgr.NewFCROfferMgr(),
			NodeOfferMap:              make(map[string]([]cidoffer.CIDOffer)),
			NodeOfferMapLock:          sync.Mutex{},
			AcknowledgementMap:        make(map[string](map[string]DHTAcknowledgement)),
			AcknowledgementMapLock:    sync.RWMutex{},
			GroupOfferGatewayIDs:      []nodeid.NodeID{},
		}
	})
	return instance
}
