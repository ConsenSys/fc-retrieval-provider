package provider

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
)

// Provider configuration
type Provider struct {
	Conf            *viper.Viper
	GatewayCommPool *fcrtcpcomms.CommunicationPool
	offers     			map[string]([]*cidoffer.CidGroupOffer)
}

// NewProvider returns new provider
func NewProvider(conf *viper.Viper) *Provider {
	gatewayCommPool := fcrtcpcomms.NewCommunicationPool()
	return &Provider{
		Conf:            conf,
		GatewayCommPool: &gatewayCommPool,
		offers:          make(map[string]([]*cidoffer.CidGroupOffer)),
	}
}

// SendMessageToGateway to gateway
func SendMessageToGateway(message *fcrmessages.FCRMessage, nodeID *nodeid.NodeID, gCommPool *fcrtcpcomms.CommunicationPool) error {
	gComm, err := gCommPool.GetConnForRequestingNode(nodeID)
	if err != nil {
		logging.Error("Connection issue: %v", err)
		if gComm != nil {
			logging.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		logging.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return err
	}
	gComm.CommsLock.Lock()
	defer gComm.CommsLock.Unlock()
	logging.Info("Send message to: %v, message: %v", nodeID.ToString(), message)
	err = fcrtcpcomms.SendTCPMessage(
		gComm.Conn,
		message,
		30000)
	if err != nil {
		logging.Error("Message not sent: %v", err)
		if gComm != nil {
			logging.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		logging.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return err
	}
	return nil
}

// AppendOffer to offers map
func (p *Provider) AppendOffer(gatewayID *nodeid.NodeID, offer *cidoffer.CidGroupOffer) {
	var offers = p.offers[strings.ToLower(gatewayID.ToString())]
	p.offers[strings.ToLower(gatewayID.ToString())] = append(offers, offer)
}

// GetOffers from offers map
func (p *Provider) GetOffers(gatewayID *nodeid.NodeID) ([]*cidoffer.CidGroupOffer) {
	return p.offers[strings.ToLower(gatewayID.ToString())]
}