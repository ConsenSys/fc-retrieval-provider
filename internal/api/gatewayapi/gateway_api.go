package gatewayapi

import (
	"net"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
)

// StartGatewayAPI starts the TCP API as a separate go routine.
func StartGatewayAPI(settings settings.AppSettings) error {
	// Start server
	ln, err := net.Listen("tcp", ":"+settings.BindGatewayAPI)
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
			log.Info("Incoming connection from gateway at :%s", conn.RemoteAddr())
			go handleIncomingGatewayConnection(conn)
		}
	}(ln)
	log.Info("Listening on %s for connections from Gateways", settings.BindGatewayAPI)

	return nil
}

func handleIncomingGatewayConnection(conn net.Conn) {
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
		if err == nil {
			if message.MessageType == fcrmessages.GatewayDHTDiscoverRequestType {
				err = handleSingleCIDOffersPublishRequest(conn, message)
				if err != nil && fcrtcpcomms.IsTimeoutError(err) {
					// Error in tcp communication, drop the connection
					log.Error1(err)
					return
				}
				continue
			}
			// Message is invalid.
			err = fcrtcpcomms.SendInvalidMessage(conn, settings.DefaultTCPInactivityTimeout)
			if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
				// Error in tcp communication, drop the connection.
				log.Error1(err)
				return
			}
		}
	}
}
