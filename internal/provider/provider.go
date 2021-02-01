package provider

import (
	"time"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/gateway"
	"github.com/spf13/viper"
)

// Provider configuration
type Provider struct {
	Conf 						*viper.Viper
	GatewayCommPool *fcrtcpcomms.CommunicationPool
}

// NewProvider returns new provider
func NewProvider(conf *viper.Viper) *Provider {
	gatewayCommPool := fcrtcpcomms.NewCommunicationPool()
	return &Provider{
		Conf: 						conf,
		GatewayCommPool: 	&gatewayCommPool,
	}
}

// Start a provider
func (provider *Provider) Start() {
	provider.greet()
	provider.registration()
	provider.loop()
}

// Greeting
func (provider *Provider) greet() {
	scheme := provider.Conf.GetString("SERVICE_SCHEME")
	host := provider.Conf.GetString("SERVICE_HOST")
	port := provider.Conf.GetString("SERVICE_PORT")
	log.Info("Provider started at %s://%s:%s", scheme, host, port)
}

// Register the provider
func (provider *Provider) registration() {
	err := RegisterProvider(provider.Conf)
	if err != nil {
		log.Error("Provider not registered: %v", err)
		//TODO graceful exit
	}
}

// Start infinite loop
func (provider *Provider) loop() {
	for {
		gateways, err := GetRegisteredGateways(provider.Conf)
		if err != nil {
			log.Error("Unable to get registered gateways: %v", err)
			//TODO graceful exit
		}
		for _, gw := range gateways {
			message := generateDummyMessage()
			log.Info("Message: %v", message)
			gatewayID, err := nodeid.NewNodeIDFromString(gw.NodeID)
			if err != nil {
				log.Error("Error with nodeID %v: %v", gw.NodeID, err)
				continue
			}
			provider.GatewayCommPool.RegisterNodeAddress(gatewayID, gw.NetworkGatewayInfo)
			gateway.SendMessage(message, gatewayID, provider.GatewayCommPool)
		}
		time.Sleep(25 * time.Second)
	}
}
