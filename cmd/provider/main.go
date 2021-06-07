/*
Package main - program entry point for a Retrieval Provider node.

Retrieval Provider is a type of nodes in FileCoin blockchain network, which serves purpose of being a way to
communicate with a Storage Miner.

Retrieval Provider is used by Retrieval Gateways in order to get their files back from the particular Storage Miner
in the network.
*/
package main

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
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrp2pserver"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrregistermgr"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrrestserver"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"

	"github.com/ConsenSys/fc-retrieval-provider/config"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/adminapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/clientapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/gatewayapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/providerapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util"
)

// Start Provider service
func main() {
	conf := config.NewConfig()
	appSettings := config.Map(conf)
	logging.Init(conf)
	logging.Info("Filecoin Provider Start-up: Started")

	logging.Info("Settings: %+v", appSettings)

	// Initialise the provider's core structure
	c := core.GetSingleInstance(&appSettings)

	// Initialise a register manager
	c.RegisterMgr = fcrregistermgr.NewFCRRegisterMgr(appSettings.RegisterAPIURL, false, true, 10*time.Second)

	// Start register manager's routine
	c.RegisterMgr.Start()

	// Create REST Server
	c.RESTServer = fcrrestserver.NewFCRRESTServer(
		[]string{appSettings.BindAdminAPI, appSettings.BindRestAPI})
	// Add handlers to the REST Server
	c.RESTServer.
		// client api
		AddHandler(appSettings.BindRestAPI, fcrmessages.ClientDHTOfferAckRequestType, clientapi.HandleClientDHTOfferAckRequest).
		// admin api
		AddHandler(appSettings.BindAdminAPI, fcrmessages.ProviderAdminInitialiseKeyRequestType, adminapi.HandleProviderAdminInitialiseKeyRequest).
		AddHandler(appSettings.BindAdminAPI, fcrmessages.ProviderAdminInitialiseKeyRequestV2Type, adminapi.HandleProviderAdminInitialiseKeyRequestV2).
		AddHandler(appSettings.BindAdminAPI, fcrmessages.ProviderAdminGetPublishedOfferRequestType, adminapi.HandleProviderAdminGetPublishedOfferRequest).
		AddHandler(appSettings.BindAdminAPI, fcrmessages.ProviderAdminPublishDHTOfferRequestType, adminapi.HandleProviderAdminPublishDHTOfferRequest).
		AddHandler(appSettings.BindAdminAPI, fcrmessages.ProviderAdminPublishGroupOfferRequestType, adminapi.HandleProviderAdminPublishGroupOfferRequest).
		AddHandler(appSettings.BindAdminAPI, fcrmessages.ProviderAdminForceRefreshRequestType, adminapi.HandleProviderAdminForceRefreshRequest)

	// Start REST Server
	err := c.RESTServer.Start()
	if err != nil {
		logging.Error("Error starting REST server: %s", err.Error())
		return
	}

	// Create P2P Server
	c.P2PServer = fcrp2pserver.NewFCRP2PServer(
		[]string{appSettings.BindGatewayAPI},
		c.RegisterMgr,
		appSettings.TCPInactivityTimeout)

	// Add handlers and requesters to the P2P Server
	c.P2PServer.
		// gateway api
		AddHandler(appSettings.BindGatewayAPI, fcrmessages.GatewayListDHTOfferRequestType, gatewayapi.HandleGatewayListDHTOfferRequest).
		AddHandler(appSettings.BindGatewayAPI, fcrmessages.GatewayNotifyProviderGroupCIDOfferSupportedRequestType, gatewayapi.HandleGatewayNotifyProviderGroupCIDOfferSupportRequest).
		// provider api
		AddRequester(fcrmessages.ProviderPublishGroupOfferRequestType, providerapi.RequestProviderPublishGroupOffer).
		AddRequester(fcrmessages.ProviderPublishDHTOfferRequestType, providerapi.RequestProviderPublishDHTOffer)

	// Start P2P Server
	err = c.P2PServer.Start()
	if err != nil {
		logging.Error("Error starting P2P server: %s", err.Error())
		return
	}

	// Configure what should be called if Control-C is hit.
	util.SetUpCtrlCExit(gracefulExit)

	logging.Info("Filecoin Provider Start-up Complete")

	// Wait forever.
	select {}
}

// gracefulExit handles exit
func gracefulExit() {
	logging.Info("Filecoin Provider Shutdown: Start")

	// TODO: Add shutdown process
	logging.Error("graceful shutdown code not written yet!")

	logging.Info("Filecoin Provider Shutdown: Completed")
}
