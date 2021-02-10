package dummy

import (
	"math/rand"
	"time"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cid"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
)

// GenerateDummyMessage is cool
func GenerateDummyMessage() *fcrmessages.FCRMessage {
	providerID, _ := nodeid.NewRandomNodeID()
	expiryDate := time.Now().Local().Add(time.Hour * time.Duration(24)).Unix()
	contentID, _ := cid.NewRandomContentID()
	pieceCIDs := []cid.ContentID{*contentID}

	cidOffer := cidoffer.CidGroupOffer{
		NodeID:     providerID,
		Cids:       pieceCIDs,
		Price:      42,
		Expiry:     expiryDate,
		QoS:        42,
		MerkleTrie: nil,
		Signature:  "",
	}

	dummyMessage, err := fcrmessages.EncodeProviderPublishGroupCIDRequest(
		rand.Int63n(100000),
		&cidOffer,
	)
	if err != nil {
		log.Error("Error when encoding message: %v", err)
	}
	return dummyMessage
}
