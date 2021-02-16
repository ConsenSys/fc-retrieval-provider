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

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrtcpcomms"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-gateway/pkg/nodeid"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
	"github.com/ConsenSys/fc-retrieval-provider/internal/register"
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
			if message.MessageType == fcrmessages.ProviderPublishGroupCIDRequestType {
				err = handleProviderPublishGroupCID(conn, p, message)
				if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
					// Error in tcp communication, drop the connection.
					logging.Error1(err)
					return
				}
				continue
			} else if message.MessageType == fcrmessages.ProviderAdminGetGroupCIDRequestType {
				err = handleProviderGetGroupCID(conn, p, message)
				if err != nil && !fcrtcpcomms.IsTimeoutError(err) {
					// Error in tcp communication, drop the connection.
					logging.Error1(err)
					return
				}
				continue
			}
		}

		// Message is invalid.
		fcrtcpcomms.SendInvalidMessage(conn, tcpInactivityTimeout)
	}
}

func handleProviderPublishGroupCID(conn net.Conn, p *provider.Provider, message *fcrmessages.FCRMessage) error {
	logging.Info("handleProviderPublishGroupCID: %+v", message)
	gateways, err := register.GetRegisteredGateways(p)
	if err != nil {
		logging.Error("Error with get registered gateways %v", err)
		return err
	}
	for _, gw := range gateways {
		gatewayID, err := nodeid.NewNodeIDFromString(gw.NodeID)
		if err != nil {
			logging.Error("Error with nodeID %v: %v", gw.NodeID, err)
			continue
		}
		err = provider.SendMessageToGateway(message, gatewayID, p.GatewayCommPool)
		if err != nil {
			logging.Error("Error with send message: %v", err)
			continue
		}
		_, offer, _ := fcrmessages.DecodeProviderPublishGroupCIDRequest(message)
		p.AppendOffer(gatewayID, offer)
	}
	return nil
}

func handleProviderGetGroupCID(conn net.Conn, p *provider.Provider, message *fcrmessages.FCRMessage) error {
	logging.Info("handleProviderGetGroupCID: %+v", message)
	gatewayID, err1 := fcrmessages.DecodeProviderAdminGetGroupCIDRequest(message)
	if err1 != nil {
		logging.Info("Provider get group cid request fail to decode request.")
		return err1
	}
	offers := p.GetOffers(gatewayID)
	message, err2 := fcrmessages.EncodeProviderAdminGetGroupCIDResponse(
		gatewayID,
		(len(offers) > 0),
		offers,
		nil,
		nil,
		nil,
	)
	if err2 != nil {
		logging.Info("Provider get group cid request fail to encode response.")
		return err2
	}
	tcpInactivityTimeout := time.Duration(p.Conf.GetInt("TCP_INACTIVITY_TIMEOUT")) * time.Millisecond
	return fcrtcpcomms.SendTCPMessage(conn, message, tcpInactivityTimeout)
}