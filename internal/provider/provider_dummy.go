package provider

// import (
// 	"math/rand"
// 	"time"

// 	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cid"
// 	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
// 	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
// 	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
// )

// func generateDummyMessage() *fcrmessages.FCRMessage {
// 	providerID, _ := nodeid.NewRandomNodeID()
// 	expiryDate := time.Now().Local().Add(time.Hour * time.Duration(24)).Unix()
// 	contentID, _ := cid.NewRandomContentID()
// 	pieceCIDs := []cid.ContentID{*contentID}
// 	dummyMessage, err := fcrmessages.EncodeProviderPublishGroupCIDRequest(
// 		rand.Int63n(100000),
// 		providerID,
// 		42,
// 		expiryDate,
// 		42,
// 		pieceCIDs,
// 	)
// 	if err != nil {
// 		log.Error("Error when encoding message: %v", err)
// 	}
// 	return dummyMessage
// }
