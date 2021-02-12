package register

import (
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
)

// RegisterProvider
func RegisterProvider(p *provider.Provider) error {
	reg := register.ProviderRegister{
		NodeID: 					"101112131415161718191A1B1C1D1E1F202122232425262728292A2B2C2D2E2F",
		Address: 					"f0121345",
		NetworkInfo:    	"localhost:9030",
		RegionCode:     	"US",
		RootSigningKey: 	"0xABCDE123456789",
		SigingKey:      	"0x987654321EDCBA",
	}
	err := register.RegisterProvider(p.Conf.GetString("REGISTER_API_URL"), reg)
	if err != nil {
		logging.Error("Provider not registered: %v", err)
	}
	return nil
}

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