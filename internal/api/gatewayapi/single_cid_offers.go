package gatewayapi

import (
	"errors"
	"net"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
)

func handleSingleCIDOffersPublishRequest(conn net.Conn, request *fcrmessages.FCRMessage) error {
	// No implementation
	return errors.New("no implementation yet")
}
