package main

import (
	"strings"
	"time"

	log "github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/register"
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

	// Start admin API first
	err := adminapi.StartAdminRestAPI(settings)
	if err != nil {
		log.Error("Error starting admin tcp server: %s", err.Error())
		return
	}
	// Configure what should be called if Control-C is hit.
	util.SetUpCtrlCExit(gracefulExit)

	// Wait until private key is set, check every 1 second
	for {
		if c.ProviderPrivateKey != nil {
			break
		}
		time.Sleep(time.Second)
	}
	log.Info("Provider private key set.")

	// Get all registerd Gateways
	gateways, err := register.GetRegisteredGateways(settings.RegisterAPIURL)
	if err != nil {
		log.Error("Unable to get registered gateways: %v", err)
	}
	c.RegisteredGatewaysMapLock.Lock()
	log.Info("All registered gateways: %+v", gateways)
	for _, gateway := range gateways {
		log.Info("Add to registered gateways map: nodeID=%+v", gateway.NodeID)
		c.RegisteredGatewaysMap[strings.ToLower(gateway.NodeID)] = &gateway
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
