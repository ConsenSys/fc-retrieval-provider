package core

import (
	"sync"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/internal/offers"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

const (
	protocolVersion   = 1 // Main protocol version
	protocolSupported = 1 // Alternative protocol version
)

// DHTAcknowledgement stores the acknowledgement of a single cid offer
type DHTAcknowledgement struct {
	Msg    fcrmessages.ProviderDHTPublishGroupCIDRequest
	MsgAck fcrmessages.ProviderDHTPublishGroupCIDAck
}

// Core holds the main data structure for the whole provider
type Core struct {
	ProtocolVersion   int32
	ProtocolSupported []int32

	// ProviderID of this provider
	ProviderID *nodeid.NodeID

	// Provider Private Key of this provider
	ProviderPrivateKey *fcrcrypto.KeyPair

	// ProviderPrivateKeyVersion is the key version number of the private key.
	ProviderPrivateKeyVersion *fcrcrypto.KeyVersion

	// RegisteredGatewaysMap stores mapping from gateway id (big int in string repr) to its registration info
	RegisteredGatewaysMap     map[string]string //TODO: Need to wait for an PR to introduce rego info
	RegisteredGatewaysMapLock sync.RWMutex

	GatewayCommPool *fcrtcpcomms.CommunicationPool

	// Offers offered by this provider, it is threadsafe.
	Offers *offers.Offers

	// Acknowledgement for every single cid offer sent
	AcknowledgementMap     map[string](map[string]DHTAcknowledgement)
	AcknowledgementMapLock sync.RWMutex
}

// Single instance of the provider
var instance *Core
var doOnce sync.Once

// GetSingleInstance returns the single instance of the provider
func GetSingleInstance(confs ...*settings.AppSettings) *Core {
	doOnce.Do(func() {
		if len(confs) == 0 {
			log.ErrorAndPanic("No settings supplied to Gateway start-up")
		}
		if len(confs) != 1 {
			log.ErrorAndPanic("More than one sets of settings supplied to Gateway start-up")
		}
		conf := confs[0]

		providerPrivateKey, err := fcrcrypto.DecodePrivateKey(conf.ProviderPrivKey)
		if err != nil {
			logging.ErrorAndPanic("Error decoding Provider Private Key: %s", err)
		}
		providerID, err := nodeid.NewNodeIDFromString(conf.ProviderID)
		if err != nil {
			logging.ErrorAndPanic("Error decoding node id: %s", err)
		}

		providerPrivateKeyVersion := fcrcrypto.DecodeKeyVersion(conf.ProviderKeyVersion)

		instance = &Core{
			ProtocolVersion:   protocolVersion,
			ProtocolSupported: []int32{protocolVersion, protocolSupported},

			ProviderID:                providerID,
			ProviderPrivateKey:        providerPrivateKey,
			ProviderPrivateKeyVersion: providerPrivateKeyVersion,

			RegisteredGatewaysMap:     make(map[string]string), //TODO, wait for PR
			RegisteredGatewaysMapLock: sync.RWMutex{},

			Offers: offers.GetSingleInstance(),

			AcknowledgementMap:     make(map[string](map[string]DHTAcknowledgement)),
			AcknowledgementMapLock: sync.RWMutex{},
		}
		// TODO, wait for PR
		// instance.GatewayCommPool = fcrtcpcomms.NewCommunicationPool(instance.RegisteredGatewaysMap, &instance.RegisteredGatewaysMapLock)
	})
	return instance
}
