package main

import (
	"strings"
	"time"

	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"
	_ "github.com/joho/godotenv/autoload"

	"github.com/ConsenSys/fc-retrieval-provider/config"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/adminapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/clientapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/api/gatewayapi"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
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

	// Get all registerd Gateways
	go updateRegisteredGateways(appSettings, c)

	// Configure what should be called if Control-C is hit.
	util.SetUpCtrlCExit(gracefulExit)

	logging.Info("Filecoin Provider Start-up Complete")

	// Wait forever.
	select {}
}

func updateRegisteredGateways(appSettings settings.AppSettings, c *core.Core) {
	for {
		logging.Debug("Update registered providers")
		gateways, err := register.GetRegisteredGateways(appSettings.RegisterAPIURL)
		if err != nil {
			logging.Error("Error in getting registered gateways: %s", err.Error())
		} else {
			// Check if nothing is changed.
			update := false
			c.RegisteredGatewaysMapLock.RLock()
			if len(gateways) != len(c.RegisteredGatewaysMap) {
				update = true
			} else {
				for _, gateway := range gateways {
					storedInfo, exist := c.RegisteredGatewaysMap[strings.ToLower(gateway.NodeID)]
					if !exist {
						update = true
						break
					} else {
						key, err := storedInfo.GetRootSigningKey()
						rootSigningKey, err2 := key.EncodePublicKey()
						key, err3 := storedInfo.GetSigningKey()
						signingKey, err4 := key.EncodePublicKey()
						if err != nil || err2 != nil || err3 != nil || err4 != nil {
							logging.Error("Error in generating key string")
							break
						}
						if gateway.Address != storedInfo.GetAddress() ||
							gateway.NetworkInfoAdmin != storedInfo.GetNetworkInfoAdmin() ||
							gateway.NetworkInfoClient != storedInfo.GetNetworkInfoClient() ||
							gateway.NetworkInfoProvider != storedInfo.GetNetworkInfoProvider() ||
							gateway.NetworkInfoGateway != storedInfo.GetNetworkInfoGateway() ||
							gateway.RegionCode != storedInfo.GetRegionCode() ||
							gateway.RootSigningKey != rootSigningKey ||
							gateway.SigningKey != signingKey {
							update = true
							break
						}
					}
				}
			}
			c.RegisteredGatewaysMapLock.RUnlock()
			if update {
				c.RegisteredGatewaysMapLock.Lock()
				c.RegisteredGatewaysMap = make(map[string]register.RegisteredNode)
				for _, gateway := range gateways {
					logging.Info("Add to registered gateways map: nodeID=%+v", gateway.NodeID)
					c.RegisteredGatewaysMap[strings.ToLower(gateway.NodeID)] = &gateway
				}
				c.RegisteredGatewaysMapLock.Unlock()
			}
		}
		// Sleep for RegisterRefreshDuration duration, refresh every RegisterRefreshDuration duration
		time.Sleep(appSettings.RegisterRefreshDuration)
	}
}

func gracefulExit() {
	logging.Info("Filecoin Provider Shutdown: Start")

	logging.Error("graceful shutdown code not written yet!")
	// TODO

	logging.Info("Filecoin Provider Shutdown: Completed")
}
