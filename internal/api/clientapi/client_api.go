package clientapi

import (
	"io/ioutil"
	"net/http"

	"github.com/ConsenSys/fc-retrieval-gateway/pkg/fcrmessages"
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ConsenSys/fc-retrieval-provider/internal/core"
	"github.com/ConsenSys/fc-retrieval-provider/internal/util/settings"
	"github.com/ant0ine/go-json-rest/rest"
)

// StartClientRestAPI starts the REST API as a separate go routine.
// Any start-up errors are returned.
func StartClientRestAPI(settings settings.AppSettings) error {
	// Start the REST API and block until the error code is set.
	errChan := make(chan error, 1)
	go startRestAPI(settings, errChan)
	return <-errChan
}

func startRestAPI(settings settings.AppSettings, errChannel chan<- error) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/v1", msgRouter),
	)
	if err != nil {
		log.Error(err.Error())
		errChannel <- err
		return
	}

	log.Info("Running REST API on: %s", settings.BindRestAPI)
	api.SetApp(router)
	errChannel <- nil
	log.Error(http.ListenAndServe(":"+settings.BindRestAPI, api.MakeHandler()).Error())
	panic("Error binding")
}

func msgRouter(w rest.ResponseWriter, r *rest.Request) {
	// Get core structure
	c := core.GetSingleInstance()

	log.Trace("Received request via /v1 API")
	content, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Error("Error reading request: %s.", err.Error())
		rest.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	if len(content) == 0 {
		log.Error("Error empty request")
		rest.Error(w, "Error empty request", http.StatusBadRequest)
		return
	}
	request, err := fcrmessages.FCRMsgFromBytes(content)
	if err != nil {
		log.Error("Failed to decode payload: %s.", err.Error())
		rest.Error(w, "Failed to decode payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Only process the rest of the message if the protocol version is understood.
	if request.ProtocolVersion != c.ProtocolVersion {
		// Check to see if the client supports the gateway's preferred version
		for _, clientProvVer := range request.ProtocolSupported {
			if clientProvVer == c.ProtocolVersion {
				// Request the client switch to this protocol version
				// TODO what can we get from request object?
				log.Info("Requesting client (TODO) switch protocol versions from %d to %d", request.ProtocolVersion, c.ProtocolVersion)
				response, _ := fcrmessages.EncodeProtocolChangeResponse(c.ProtocolVersion)
				w.WriteJson(response)
				return
			}
		}

		// Go through the protocol versions supported by the client and the
		// gateway to search for any common version, prioritising
		// the gateway preference over the client preference.
		for _, clientProvVer := range request.ProtocolSupported {
			for _, gatewayProtVer := range c.ProtocolSupported {
				if clientProvVer == gatewayProtVer {
					// When we support more than one version of the protocol, this code will change the gateway
					// to using the other (common version)
					log.Error("Not implemented yet")
					panic("Multiple protocol versions not implemented yet")
				}
			}
		}
		// No common protocol versions supported.
		// TODO what can we get from request object?
		log.Warn("Client Request: Unsupported protocol version(s): %d", request.ProtocolVersion)
		response, _ := fcrmessages.EncodeProtocolMismatchResponse()
		w.WriteJson(response)
		return
	}

	switch request.GetMessageType() {
	case fcrmessages.ClientCIDGroupPublishDHTAckRequestType:
		handleClientCIDGroupPublishDHTAckRequest(w, request)
	default:
		log.Warn("Client Request: Unknown message type: %d", request.GetMessageType())
		rest.Error(w, "Unknown message type", http.StatusBadRequest)
	}
}
