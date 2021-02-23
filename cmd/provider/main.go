package main

import (
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/register"
	_ "github.com/joho/godotenv/autoload"

	"github.com/ConsenSys/fc-retrieval-provider/config"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/adminapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/clientapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/gatewayapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util"
)

// Start Provider service
func main() {
	conf := config.NewConfig()
	settings := config.Map(conf)
	log.Init(conf)
	log.Info("Filecoin Provider Start-up: Started")

	log.Info("Settings: %+v", settings)

	// Initialise the provider's core structure
	c := core.GetSingleInstance(&settings)

	// A few things TODO:
	// 1. Loop and check if this provider is registered, exit loop when this provider is registered by the admin
	// 2. Loop and check if this provider has a key provided, exit loop when a key has been provided by the admin
	// 3. (Concurrently), get all registered gateways every 10 seconds, this can be done by a seperate go-routine.
	// Register provider, TO BE REPLACED, as the registration is done in admin
	providerReg := register.ProviderRegister{
		NodeID:             settings.ProviderID,
		Address:            settings.ProviderAddress,
		RootSigningKey:     settings.ProviderRootSigningKey,
		SigningKey:         settings.ProviderSigningKey,
		NetworkGatewayInfo: settings.NetworkGatewayInfo,
		NetworkClientInfo:  settings.NetworkClientInfo,
		NetworkAdminInfo:   settings.NetworkAdminInfo,
		RegionCode:         settings.ProviderRegionCode,
	}
	providerReg.RegisterProvider(settings.RegisterAPIURL)

	// Get all registerd Gateways
	gateways, err := register.GetRegisteredGateways(settings.RegisterAPIURL)
	if err != nil {
		log.Error("Unable to get registered gateways: %v", err)
	}
	c.RegisteredGatewaysMapLock.Lock()
	log.Info("All registered gateways: %+v", gateways)
	for _, gateway := range gateways {
		c.RegisteredGatewaysMap[gateway.NodeID] = &gateway
	}
	c.RegisteredGatewaysMapLock.Unlock()

	err = clientapi.StartClientRestAPI(settings)
	if err != nil {
		log.Error("Error starting client rest server: %s", err.Error())
		return
	}

	err = gatewayapi.StartGatewayAPI(settings)
	if err != nil {
		log.Error("Error starting gateway tcp server: %s", err.Error())
	}

	err = adminapi.StartAdminRestAPI(settings)
	if err != nil {
		log.Error("Error starting admin tcp server: %s", err.Error())
		return
	}
	// Configure what should be called if Control-C is hit.
	util.SetUpCtrlCExit(gracefulExit)

	log.Info("Filecoin Provider Start-up Complete")

	// Wait forever.
	select {}
}

func gracefulExit() {
	log.Info("Filecoin Provider Shutdown: Start")

	log.Error("graceful shutdown code not written yet!")
	// TODO

	log.Info("Filecoin Provider Shutdown: Completed")
}
