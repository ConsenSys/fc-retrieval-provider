package gatewayapi

import (
	"net"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
)

func handleSingleCIDOffersPublishRequest(conn net.Conn, request *fcrmessages.FCRMessage) error {
	log.Error("Not implemented.")

	// TODO: Verifying & storing acknowledgements, will need gatewayID
	return nil
}
