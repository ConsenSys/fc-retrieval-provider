package provider

import (
	"time"

	"github.com/spf13/viper"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"

	"github.com/ConsenSys/fc-retrieval-provider/internal/dummy"
	"github.com/ConsenSys/fc-retrieval-provider/internal/gateway"
)

// Provider configuration
type Provider struct {
	Conf            *viper.Viper
	GatewayCommPool *fcrtcpcomms.CommunicationPool
}

// NewProvider returns new provider
func NewProvider(conf *viper.Viper) *Provider {
	gatewayCommPool := fcrtcpcomms.NewCommunicationPool()
	return &Provider{
		Conf:            conf,
		GatewayCommPool: &gatewayCommPool,
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
	reg := register.ProviderRegister{
		Address:        provider.Conf.GetString("PROVIDER_ADDRESS"),
		NetworkInfo:    provider.Conf.GetString("PROVIDER_NETWORK_INFO"),
		RegionCode:     provider.Conf.GetString("PROVIDER_REGION_CODE"),
		RootSigningKey: provider.Conf.GetString("PROVIDER_ROOT_SIGNING_KEY"),
		SigingKey:      provider.Conf.GetString("PROVIDER_SIGNING_KEY"),
	}

	err := register.RegisterProvider(provider.Conf.GetString("REGISTER_API_URL"), reg)
	if err != nil {
		log.Error("Provider not registered: %v", err)
	}
}

// Start infinite loop
func (provider *Provider) loop() {
	for {
		gateways, err := register.GetRegisteredGateways(provider.Conf.GetString("REGISTER_API_URL"))
		if err != nil {
			log.Error("Unable to get registered gateways: %v", err)
		}
		for _, gw := range gateways {
			message := dummy.GenerateDummyMessage()
			log.Info("Message: %v", message)
			gatewayID, err := nodeid.NewNodeIDFromString(gw.NodeID)
			if err != nil {
				log.Error("Error with nodeID %v: %v", gw.NodeID, err)
				continue
			}
			provider.GatewayCommPool.RegisterNodeAddress(gatewayID, gw.NetworkProviderInfo)
			gateway.SendMessage(message, gatewayID, provider.GatewayCommPool)
		}
		time.Sleep(25 * time.Second)
	}
}
