package adminapi

/*
 * Copyright 2020 ConsenSys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"fmt"
	"net"
	"time"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
)

// StartAdminAPI starts the TCP API as a separate go routine.
func StartAdminAPI(p *provider.Provider) error {
	// Start server
	bindAdminApi := p.Conf.GetString("SERVICE_PORT")
	ln, err := net.Listen("tcp", ":"+bindAdminApi)
	if err != nil {
		return err
	}
	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				logging.Error1(err)
				continue
			}
			logging.Info("Incoming connection from admin client at :%s", conn.RemoteAddr())
			go handleIncomingAdminConnection(conn, p)
		}
	}(ln)
	logging.Info("Listening on %s for connections from admin clients", bindAdminApi)
	return nil
}

func handleIncomingAdminConnection(conn net.Conn, p *provider.Provider) {
	// Close connection on exit.
	defer conn.Close()

	// Loop until error occurs and connection is dropped.
	tcpInactivityTimeout := time.Duration(p.Conf.GetInt("TCP_INACTIVITY_TIMEOUT")) * time.Millisecond
	for {
		message, err := fcrtcpcomms.ReadTCPMessage(conn, tcpInactivityTimeout)
		if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
			// Error in tcp communication, drop the connection.
			logging.Error1(err)
			return
		}
		// Respond to requests for a client's reputation.
		if err == nil {
			fmt.Printf("Message: %+v\n", message)
			// if message.MessageType == fcrmessages.AdminGetReputationChallengeType {
			// 	err = handleAdminGetReputationChallenge(conn, message)
			// 	if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
			// 		// Error in tcp communication, drop the connection.
			// 		logging.Error1(err)
			// 		return
			// 	}
			// 	continue
			// } else if message.MessageType == fcrmessages.AdminSetReputationChallengeType {
			// 	err = handleAdminSetReputationChallenge(conn, message)
			// 	if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
			// 		// Error in tcp communication, drop the connection.
			// 		logging.Error1(err)
			// 		return
			// 	}
			// 	continue
			// } else if message.MessageType == fcrmessages.AdminAcceptKeyChallengeType {
			// 	err = handleAdminAcceptKeysChallenge(conn, message)
			// 	if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
			// 		// Error in tcp communication, drop the connection.
			// 		logging.Error1(err)
			// 		return
			// 	}
			// 	continue
			// }
		}

		// Message is invalid.
		fcrtcpcomms.SendInvalidMessage(conn, tcpInactivityTimeout)
	}
}

// // GetConnForRequestingAdminClient returns the connection for sending request to an admin client with given id.
// // It will reuse any active connection.
// func GetConnForRequestingAdminClient(gatewayID nodeid.NodeID, g *gateway.Gateway) (*gateway.CommunicationChannel, error) {
// 	// Check if there is an active connection.
// 	g.ActiveGatewaysLock.RLock() // TODO: Check this - Will need to add active admin connections to core structure
// 	gComm := g.ActiveGateways[gatewayID.ToString()]
// 	g.ActiveGatewaysLock.RUnlock()
// 	if gComm == nil {
// 		// No active connection, connect to peer.
// 		g.GatewayAddressMapLock.RLock()
// 		conn, err := net.Dial("tcp", g.GatewayAddressMap[gatewayID.ToString()])
// 		g.GatewayAddressMapLock.RUnlock()
// 		if err != nil {
// 			return nil, err
// 		}
// 		gComm = &gateway.CommunicationChannel{
// 			CommsLock: sync.RWMutex{},
// 			Conn:      conn}
// 		if gateway.RegisterGatewayCommunication(&gatewayID, gComm) != nil {
// 			conn.Close()
// 			return nil, err
// 		}
// 	}
// 	return gComm, nil
// }
