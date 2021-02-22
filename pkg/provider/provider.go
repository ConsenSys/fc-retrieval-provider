package provider

import (
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/register"
)

// Provider configuration
type Provider struct {
	Conf            		*viper.Viper
	GatewayCommPool 		*fcrtcpcomms.CommunicationPool
	ProtocolVersion			int32
	ProtocolSupported		[]int32
	Offers     					map[string]([]*cidoffer.CidGroupOffer)
}

// Single instance of the gateway
var instance *Provider
var doOnce sync.Once

// GetSingleInstance returns the single instance of the provider
func GetSingleInstance(confs ...*viper.Viper) *Provider {
	doOnce.Do(func() {
		conf := getConf(confs)
		protocolVersion := conf.GetInt32("PROTOCOL_VERSION")
		protocolSupported := conf.GetInt32("PROTOCOL_SUPPORTED")
		gatewayCommPool := fcrtcpcomms.NewCommunicationPool(make(map[string]register.RegisteredNode), &sync.RWMutex{})
		instance = &Provider{
			Conf:            		conf,
			ProtocolVersion:		protocolVersion,
			ProtocolSupported:	[]int32{protocolVersion, protocolSupported},
			GatewayCommPool: 		gatewayCommPool,
			Offers:          		make(map[string]([]*cidoffer.CidGroupOffer)),
		}
	})
	return instance
}

func getConf(confs []*viper.Viper) (*viper.Viper) {
	if len(confs) == 0 {
		logging.ErrorAndPanic("No settings supplied to Gateway start-up")
	}
	if len(confs) != 1 {
		logging.ErrorAndPanic("More than one sets of settings supplied to Gateway start-up")
	}
	return confs[0]
}

// SendMessageToGateway to gateway
func (p *Provider) SendMessageToGateway(message *fcrmessages.FCRMessage, nodeID *nodeid.NodeID) (
	*fcrmessages.FCRMessage,
	error,
) {
	tcpInactivityTimeout := time.Duration(p.Conf.GetInt("TCP_INACTIVITY_TIMEOUT")) * time.Millisecond
	gCommPool := p.GatewayCommPool
	gComm, err := gCommPool.GetConnForRequestingNode(nodeID, fcrtcpcomms.AccessFromProvider)
	if err != nil {
		logging.Error("Connection issue: %v", err)
		if gComm != nil {
			logging.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		logging.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return nil, err
	}
	gComm.CommsLock.Lock()
	defer gComm.CommsLock.Unlock()
	logging.Info("Send message to: %v, message: %v", nodeID.ToString(), message)
	err = fcrtcpcomms.SendTCPMessage(
		gComm.Conn,
		message,
		tcpInactivityTimeout)
	if err != nil {
		logging.Error("Message not sent: %v", err)
		if gComm != nil {
			logging.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		logging.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return nil, err
	}
	response, err := fcrtcpcomms.ReadTCPMessage(gComm.Conn, tcpInactivityTimeout)
	if err != nil && fcrtcpcomms.IsTimeoutError(err) {
		// Timeout can be ignored. Since this message can expire.
		return nil, nil
	} else if err != nil {
		logging.Error("Message not sent: %v", err)
		if gComm != nil {
			logging.Debug("Closing connection ...")
			gComm.Conn.Close()
		}
		logging.Debug("Removing connection from pool ...")
		gCommPool.DeregisterNodeCommunication(nodeID)
		return nil, err
	}
	return response, nil
}

// GetAllOffers from offers map
func (p *Provider) GetAllOffers() ([]*cidoffer.CidGroupOffer) {
	var offers []*cidoffer.CidGroupOffer
	for _, values := range p.Offers {
		for _, value := range values {
			offers = append(offers, value)
		}
	}
	return offers
}

// GetOffersByGatewayID from offers map
func (p *Provider) GetOffersByGatewayID(gatewayID *nodeid.NodeID) ([]*cidoffer.CidGroupOffer) {
	return p.Offers[strings.ToLower(gatewayID.ToString())]
}

// AppendOffer to offers map
func (p *Provider) AppendOffer(gatewayID *nodeid.NodeID, offer *cidoffer.CidGroupOffer) {
	var offers = p.Offers[strings.ToLower(gatewayID.ToString())]
	p.Offers[strings.ToLower(gatewayID.ToString())] = append(offers, offer)
}