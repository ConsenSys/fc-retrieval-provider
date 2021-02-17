package adminapi

import (
	"net"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

// StartAdminAPI starts the TCP API as a separate go routine.
func StartAdminAPI(settings settings.AppSettings) error {
	// Start server
	ln, err := net.Listen("tcp", ":"+settings.BindAdminAPI)
	if err != nil {
		return err
	}
	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Error1(err)
				continue
			}
			log.Info("Incoming connection from admin client at :%s", conn.RemoteAddr())
			go handleIncomingAdminConnection(conn)
		}
	}(ln)
	log.Info("Listening on %s for connections from admin clients", settings.BindAdminAPI)
	return nil
}

func handleIncomingAdminConnection(conn net.Conn) {
	// Close connection on exit.
	defer conn.Close()

	// Loop until error occurs and connection is dropped.
	for {
		message, err := fcrtcpcomms.ReadTCPMessage(conn, settings.DefaultTCPInactivityTimeout)
		if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
			// Error in tcp communication, drop the connection.
			log.Error1(err)
			return
		}
		// Respond to requests for a client's reputation.
		if err == nil {
			if message.MessageType == fcrmessages.AdminAcceptKeyChallengeType {
				err = handleAdminAcceptKeysChallenge(conn, message)
				if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
					// Error in tcp communication, drop the connection.
					log.Error1(err)
					return
				}
				continue
			}
			// TODO: Add message types for adding cid offer? so the provider will publish those offers?
		}

		// Message is invalid.
		fcrtcpcomms.SendInvalidMessage(conn, settings.DefaultTCPInactivityTimeout)
	}
}
