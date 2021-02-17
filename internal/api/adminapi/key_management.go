package adminapi

import (
	"errors"
	"net"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
)

func handleAdminAcceptKeysChallenge(conn net.Conn, request *fcrmessages.FCRMessage) error {
	// No implementation
	return errors.New("no implementation yet")
}
