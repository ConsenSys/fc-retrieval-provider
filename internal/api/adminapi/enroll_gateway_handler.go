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
	"strings"

	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

func handleAdminEnrollGateway(w rest.ResponseWriter, request *fcrmessages.FCRMessage, settings settings.AppSettings) {
	coreStructure := core.GetSingleInstance(&settings)
	if coreStructure.ProviderPrivateKey == nil {
		logging.Error("this provider hasn't been initialised by the admin")
		return
	}

	nodeID, address, rootSigningKey, signingKey, regionCode, networkInfoGateway, networkInfoProvider,
		networkInfoClient, networkInfoAdmin, err :=
		fcrmessages.DecodeGatewayAdminEnrollGatewayRequest(request)
	if err != nil {
		logging.Error(err.Error())
		return
	}

	newGateway := register.GatewayRegister{
		NodeID:              nodeID.ToString(),
		Address:             address,
		RootSigningKey:      rootSigningKey,
		SigningKey:          signingKey,
		RegionCode:          regionCode,
		NetworkInfoGateway:  networkInfoGateway,
		NetworkInfoProvider: networkInfoProvider,
		NetworkInfoClient:   networkInfoClient,
		NetworkInfoAdmin:    networkInfoAdmin,
	}
	ok := registerGateway(newGateway, coreStructure)
	if !ok {
		logging.Error("can not register a gateway")
		return
	}

	// Construct a message
	response, err := fcrmessages.EncodeGatewayAdminEnrollGatewayRequest(
		nodeID,
		address,
		rootSigningKey,
		signingKey,
		regionCode,
		networkInfoGateway,
		networkInfoProvider,
		networkInfoClient,
		networkInfoAdmin)
	if err != nil {
		logging.Error(err.Error())
		return
	}
	// Sign message
	if response.Sign(coreStructure.ProviderPrivateKey, coreStructure.ProviderPrivateKeyVersion) != nil {
		logging.Error("error in signing message")
		return
	}

	w.WriteJson(response)
}

func registerGateway(newGateway register.GatewayRegister, coreStructure *core.Core) (ok bool) {
	logging.Debug("Register a new gateway")
	update := false
	coreStructure.RegisteredGatewaysMapLock.RLock()
	storedInfo, exist := coreStructure.RegisteredGatewaysMap[strings.ToLower(newGateway.NodeID)]
	if !exist {
		update = true
	} else {
		key, err := storedInfo.GetRootSigningKey()
		rootSigningKey, err2 := key.EncodePublicKey()
		key, err3 := storedInfo.GetSigningKey()
		signingKey, err4 := key.EncodePublicKey()
		if err != nil || err2 != nil || err3 != nil || err4 != nil {
			logging.Error("Error in generating key string")
			return false
		}
		if newGateway.Address != storedInfo.GetAddress() ||
			newGateway.NetworkInfoAdmin != storedInfo.GetNetworkInfoAdmin() ||
			newGateway.NetworkInfoClient != storedInfo.GetNetworkInfoClient() ||
			newGateway.NetworkInfoProvider != storedInfo.GetNetworkInfoProvider() ||
			newGateway.NetworkInfoGateway != storedInfo.GetNetworkInfoGateway() ||
			newGateway.RegionCode != storedInfo.GetRegionCode() ||
			newGateway.RootSigningKey != rootSigningKey ||
			newGateway.SigningKey != signingKey {

			update = true
		}
	}
	coreStructure.RegisteredGatewaysMapLock.RUnlock()
	if update {
		logging.Info("Add to registered gateways map: nodeID=%+v", newGateway.NodeID)
		coreStructure.RegisteredGatewaysMapLock.Lock()
		coreStructure.RegisteredGatewaysMap[strings.ToLower(newGateway.NodeID)] = &newGateway
		coreStructure.RegisteredGatewaysMapLock.Unlock()
	}
	return true
}
