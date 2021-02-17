package provider

import (
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/register"

	"github.com/ConsenSys/fc-retrieval-provider/internal/dummy"
)

// Provider configuration
type Provider struct {
	Conf            *viper.Viper
	GatewayCommPool *fcrtcpcomms.CommunicationPool
}

// NewProvider returns new provider
func NewProvider(conf *viper.Viper) *Provider {
	gatewayCommPool := fcrtcpcomms.NewCommunicationPool(make(map[string]register.RegisteredNode), &sync.RWMutex{})
	return &Provider{
		Conf:            conf,
		GatewayCommPool: gatewayCommPool,
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
		Address:            provider.Conf.GetString("PROVIDER_ADDRESS"),
		NetworkGatewayInfo: provider.Conf.GetString("PROVIDER_NETWORK_INFO"),
		RegionCode:         provider.Conf.GetString("PROVIDER_REGION_CODE"),
		RootSigningKey:     provider.Conf.GetString("PROVIDER_ROOT_SIGNING_KEY"),
		SigningKey:         provider.Conf.GetString("PROVIDER_SIGNING_KEY"),
	}

	err := reg.RegisterProvider(provider.Conf.GetString("REGISTER_API_URL"))
	if err != nil {
		log.Error("Provider not registered: %v", err)
	}
}

// SendMessageToGateway to gateway
func SendMessageToGateway(message *fcrmessages.FCRMessage, nodeID *nodeid.NodeID, gCommPool *fcrtcpcomms.CommunicationPool) error {
	gComm, err := gCommPool.GetConnForRequestingNode(nodeID, fcrtcpcomms.AccessFromProvider)
	if err != nil {
		log.Error("Connection issue: %v", err)
		if gComm != nil {
			log.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		log.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return err
	}
	gComm.CommsLock.Lock()
	defer gComm.CommsLock.Unlock()
	log.Info("Send message to: %v, message: %v", nodeID.ToString(), message)
	err = fcrtcpcomms.SendTCPMessage(
		gComm.Conn,
		message,
		30000)
	if err != nil {
		log.Error("Message not sent: %v", err)
		if gComm != nil {
			log.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		log.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return err
	}
	return nil
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
			provider.GatewayCommPool.RegisteredNodeMap[gw.NodeID] = &gw
			SendMessageToGateway(message, gatewayID, provider.GatewayCommPool)
		}
		time.Sleep(25 * time.Second)
	}
}
