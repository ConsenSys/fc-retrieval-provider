package core

import (
	"sync"

	"github.com/ConsenSys/fc-retrieval-common/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrcrypto"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-common/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	log "github.com/ConsenSys/fc-retrieval-common/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-common/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-common/pkg/register"
	"github.com/ConsenSys/fc-retrieval-provider/internal/offers"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

const (
	protocolVersion   = 1 // Main protocol version
	protocolSupported = 1 // Alternative protocol version
)

// DHTAcknowledgement stores the acknowledgement of a single cid offer
type DHTAcknowledgement struct {
	Msg    fcrmessages.FCRMessage // Original message
	MsgAck fcrmessages.FCRMessage // Original message ACK
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
	RegisteredGatewaysMap     map[string]register.RegisteredNode
	RegisteredGatewaysMapLock sync.RWMutex

	GatewayCommPool *fcrtcpcomms.CommunicationPool

	// GroupOffers sent, it is threadsafe
	GroupOffers *offers.Offers

	// SingleOffers sent, it is threadsafe
	SingleOffers *offers.Offers

	// Node to offer map
	NodeOfferMap     map[string]([]cidoffer.CidGroupOffer)
	NodeOfferMapLock sync.Mutex

	// Acknowledgement for every single cid offer sent (map from cid id -> map of gateway -> ack)
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

		providerID, err := nodeid.NewNodeIDFromString(conf.ProviderID)
		if err != nil {
			logging.ErrorAndPanic("Error decoding node id: %s", err)
		}

		instance = &Core{
			ProtocolVersion:   protocolVersion,
			ProtocolSupported: []int32{protocolVersion, protocolSupported},

			ProviderID:                providerID,
			ProviderPrivateKey:        nil,
			ProviderPrivateKeyVersion: nil,

			RegisteredGatewaysMap:     make(map[string]register.RegisteredNode),
			RegisteredGatewaysMapLock: sync.RWMutex{},

			GroupOffers:      offers.GetSingleInstance(),
			SingleOffers:     offers.GetSingleInstance(),
			NodeOfferMap:     make(map[string]([]cidoffer.CidGroupOffer)),
			NodeOfferMapLock: sync.Mutex{},

			AcknowledgementMap:     make(map[string](map[string]DHTAcknowledgement)),
			AcknowledgementMapLock: sync.RWMutex{},
		}
		instance.GatewayCommPool = fcrtcpcomms.NewCommunicationPool(instance.RegisteredGatewaysMap, &instance.RegisteredGatewaysMapLock)
	})
	return instance
}
