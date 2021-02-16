package clientapi

// Copyright (C) 2020 ConsenSys Software Inc

// Contains debug APIs

import (
	"net"
	"os"
	"time"

	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	"github.com/ant0ine/go-json-rest/rest"
)

func getTime(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(time.Now())
}

func getHostname(w rest.ResponseWriter, r *rest.Request) {
	name, err := os.Hostname()
	if err != nil {
		log.Info("Get host name1: %s", err.Error())
		return
	}

	w.WriteJson(name)
}

func getIP(w rest.ResponseWriter, r *rest.Request) {
	name, err := os.Hostname()
	if err != nil {
		log.Info("Get host name2: %s", err.Error())
		return
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Info("Lookup host: %s", err.Error())
		return
	}

	w.WriteJson(addrs)
}
