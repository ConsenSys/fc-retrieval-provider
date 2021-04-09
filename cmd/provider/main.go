package main

import (
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
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
	appSettings := config.Map(conf)
	logging.Init(conf)
	logging.Info("Filecoin Provider Start-up: Started")

	logging.Info("Settings: %+v", appSettings)

	// Initialise the provider's core structure
	core.GetSingleInstance(&appSettings)

	err := clientapi.StartClientRestAPI(appSettings)
	if err != nil {
		logging.Error("Error starting client rest server: %s", err.Error())
		return
	}

	err = gatewayapi.StartGatewayAPI(appSettings)
	if err != nil {
		logging.Error("Error starting gateway tcp server: %s", err.Error())
	}

	err = adminapi.StartAdminRestAPI(appSettings)
	if err != nil {
		logging.Error("Error starting admin tcp server: %s", err.Error())
		return
	}

	// Configure what should be called if Control-C is hit.
	util.SetUpCtrlCExit(gracefulExit)

	logging.Info("Filecoin Provider Start-up Complete")

	// Wait forever.
	select {}
}

func gracefulExit() {
	logging.Info("Filecoin Provider Shutdown: Start")

	logging.Error("graceful shutdown code not written yet!")
	// TODO

	logging.Info("Filecoin Provider Shutdown: Completed")
}
