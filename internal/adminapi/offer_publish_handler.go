package adminapi

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/cidoffer"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmerkletree"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
)

func handleProviderGetGroupCID(w rest.ResponseWriter, request *fcrmessages.FCRMessage, p *provider.Provider) {
	logging.Info("handleProviderGetGroupCID: %+v", request)
	gatewayID, err1 := fcrmessages.DecodeProviderAdminGetGroupCIDRequest(request)
	if err1 != nil {
		logging.Info("Provider get group cid request fail to decode request.")
		panic(err1)
	}
	offers := make([]*cidoffer.CidGroupOffer, 0)
	if gatewayID != nil {
		offers = p.GetOffers(gatewayID)
	} else {
		//TODO: get all offers
	}

	// TODO: fix roots, proofs and payments
	roots := make([]string, len(offers))
	proofs := make([]fcrmerkletree.FCRMerkleProof, len(offers))
	fundedPaymentChannel := make([]bool, len(offers))
	for i := 0; i < len(offers); i++ {
		roots[i] = ""
		proofs[i] = fcrmerkletree.FCRMerkleProof{}
		fundedPaymentChannel[i] = false
	}

	response, err2 := fcrmessages.EncodeProviderAdminGetGroupCIDResponse(
		gatewayID,
		len(offers) > 0,
		offers,
		roots,
		proofs,
		fundedPaymentChannel,
	)
	if err2 != nil {
		logging.Info("Provider get group cid request fail to encode response.")
		panic(err2)
	}
	w.WriteJson(response)
}