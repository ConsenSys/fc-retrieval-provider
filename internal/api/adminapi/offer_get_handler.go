package adminapi

import (
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ant0ine/go-json-rest/rest"
)

func handleProviderGetGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage, c *core.Core) {
	logging.Info("handleProviderGetGroupCID: %+v", request)
	gatewayIDs, err1 := fcrmessages.DecodeProviderAdminGetGroupCIDRequest(request)
	if err1 != nil {
		logging.Info("Provider get group cid request fail to decode request.")
		panic(err1)
	}
	logging.Info("Find offers: gatewayIDs=%+v", gatewayIDs)
	offers := make([]*cidoffer.CidGroupOffer, 0)
	c.NodeOfferMapLock.Lock()
	defer c.NodeOfferMapLock.Unlock()
	if len(gatewayIDs) > 0 {
		for _, gatewayID := range gatewayIDs {
			offs := c.NodeOfferMap[gatewayID.ToString()]
			for _, off := range offs {
				offers = append(offers, &off)
			}
		}
	} else {
		for _, values := range c.NodeOfferMap {
			for _, value := range values {
				offers = append(offers, &value)
			}
		}
	}
	logging.Info("Found offers: %+v", len(offers))

	// TODO: fix payments
	roots := make([]string, len(offers))
	fundedPaymentChannel := make([]bool, len(offers))
	for i := 0; i < len(offers); i++ {
		offer := offers[i]
		roots[i] = offer.GetMerkleRoot()
		fundedPaymentChannel[i] = false
	}

	response, err2 := fcrmessages.EncodeProviderAdminGetGroupCIDResponse(
		len(offers) > 0,
		offers,
		roots,
		fundedPaymentChannel,
	)
	if err2 != nil {
		logging.Info("Provider get group cid request fail to encode response.")
		panic(err2)
	}
	w.WriteJson(response)
}
