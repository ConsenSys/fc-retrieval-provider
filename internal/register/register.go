package register

import (
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
)

// GetRegisteredGateways gets registered gateways
func GetRegisteredGateways(p *provider.Provider) ([]register.GatewayRegister, error) {
	gateways, err := register.GetRegisteredGateways(p.Conf.GetString("REGISTER_API_URL"))
	if err != nil {
		logging.Error("Unable to get registered gateways: %v", err)
		return []register.GatewayRegister{}, err
	}
	for _, gw := range gateways {
		gatewayID, err := nodeid.NewNodeIDFromString(gw.NodeID)
		if err != nil {
			logging.Error("Error with nodeID %v: %v", gw.NodeID, err)
			continue
		}
		p.GatewayCommPool.RegisterNodeAddress(gatewayID, gw.NetworkProviderInfo)
	}
	return gateways, err
}