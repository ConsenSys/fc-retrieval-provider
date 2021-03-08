package main

import (
	"strings"
	"time"

	log "github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"
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

	err := clientapi.StartClientRestAPI(settings)
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

	// Get all registerd Gateways
	go updateRegisteredGateways(settings.RegisterAPIURL, c)

	// Configure what should be called if Control-C is hit.
	util.SetUpCtrlCExit(gracefulExit)

	log.Info("Filecoin Provider Start-up Complete")

	// Wait forever.
	select {}
}

func updateRegisteredGateways(url string, c *core.Core) {
	for {
		gateways, err := register.GetRegisteredGateways(url)
		if err != nil {
			log.Error("Error in getting registered gateways: ", err.Error())
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
							log.Error("Error in generating key string")
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
					log.Info("Add to registered gateways map: nodeID=%+v", gateway.NodeID)
					c.RegisteredGatewaysMap[strings.ToLower(gateway.NodeID)] = &gateway
				}
				c.RegisteredGatewaysMapLock.Unlock()
			}
		}
		// Sleep for 5 seconds, refresh every 5 seconds
		time.Sleep(5 * time.Second)
	}
}

func gracefulExit() {
	log.Info("Filecoin Provider Shutdown: Start")

	log.Error("graceful shutdown code not written yet!")
	// TODO

	log.Info("Filecoin Provider Shutdown: Completed")
}
