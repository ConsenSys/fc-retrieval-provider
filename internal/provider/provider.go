package provider

import (
	"fmt"
	"time"

	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/request"
	"github.com/ConsenSys/fc-retrieval-provider/internal/gateway"
	"github.com/ConsenSys/fc-retrieval-register/pkg/register"
	"github.com/spf13/viper"
)

// Provider configuration
type Provider struct {
	conf *viper.Viper
}

// NewProvider returns new provider
func NewProvider(conf *viper.Viper) *Provider {
	return &Provider{
		conf: conf,
	}
}

// Start a new provider
func Start(p *Provider) {
	scheme := p.conf.GetString("SERVICE_SCHEME")
	host := p.conf.GetString("SERVICE_HOST")
	port := p.conf.GetString("SERVICE_PORT")
	log.Info("Provider started at %s://%s:%s", scheme, host, port)
	url := p.conf.GetString("REGISTER_API_URL") + "/registers/provider"
	Registration(url, p)
	p.loop()
}

// Start infinite loop
func (p *Provider) loop() {
	url := p.conf.GetString("REGISTER_API_URL") + "/registers/gateway"

	for {
		gateways := []register.GatewayRegister{}
		request.GetJSON(url, &gateways)

		if len(gateways) == 0 {
			log.Warn("No gateways found")
		}

		for _, gw := range gateways {
			message := generateDummyMessage()
			fmt.Println(message)
			gwURL := gw.NetworkProviderInfo
			gateway.SendMessage(gwURL, message)
		}
		time.Sleep(25 * time.Second)
	}
}
