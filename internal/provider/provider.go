package provider

import (
	"math/big"
	"time"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/gateway"
	"github.com/spf13/viper"
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
	err := RegisterProvider(provider.Conf)
	if err != nil {
		log.Error("Provider not registered: %v", err)
		//TODO graceful exit
	}
}

// Start infinite loop
func (provider *Provider) loop() {
	key, _ := fcrcrypto.DecodePrivateKey("01d669ab849c3baf0491f581f498560e46d8c10571a673fd19d638389f383061a4")
	keyVersion := fcrcrypto.InitialKeyVersion()

	// My gateway address at port 8090, id 11000
	gatewayAddr := "localhost:8090"
	gatewayID, _ := nodeid.NewNodeID(big.NewInt(11000))
	provider.GatewayCommPool.RegisterNodeAddress(gatewayID, gatewayAddr)

	// My provider ID, 12000
	providerID, _ := nodeid.NewNodeID(big.NewInt(12000))

	time.Sleep(5 * time.Second)
	for {
		// Generate a random cid offer with five random cids
		cid1, _ := cid.NewRandomContentID()
		cid2, _ := cid.NewRandomContentID()
		cid3, _ := cid.NewRandomContentID()
		cid4, _ := cid.NewRandomContentID()
		cid5, _ := cid.NewRandomContentID()
		offer, err := cidoffer.NewCidGroupOffer(providerID, &[]cid.ContentID{*cid1, *cid2, *cid3, *cid4, *cid5}, 100, time.Now().Add(time.Hour*5).Unix(), 100)
		if err != nil {
			log.Error("Error in generating the cid offer.")
			continue
		}
		logging.Info("Offer created with merkle root %s", offer.MerkleRoot)
		time.Sleep(3 * time.Second)

		// Sign offer
		offer.SignOffer(func(msg interface{}) (string, error) {
			return fcrcrypto.SignMessage(key, keyVersion, msg)
		})
		logging.Info("Offer signed with signature: %s", offer.Signature)
		time.Sleep(3 * time.Second)

		// Scenario 1, wrong signature
		// offer.Signature = offer.Signature[:10] + string((offer.Signature[10] + 1)) + offer.Signature[11:]
		// logging.Info("Signature changed.")

		// Scenario 2, swap cid.
		// cid6, _ := cid.NewRandomContentID()
		// offer.Cids[0] = *cid6
		// logging.Info("CID swapped.")

		request, err := fcrmessages.EncodeProviderPublishGroupCIDRequest(1, offer)
		if err != nil {
			log.Error("Error in generating the request.")
			continue
		}

		logging.Info("Sending offer...")
		time.Sleep(3 * time.Second)
		gateway.SendMessage(request, gatewayID, provider.GatewayCommPool)
		logging.Info("Offer sent.")

		// Sleep
		time.Sleep(10 * time.Second)
	}
}
