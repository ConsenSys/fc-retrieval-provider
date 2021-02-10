package main

import (
	log "github.com/ConsenSys/fc-retrieval-gateway/pkg/logging"
	_ "github.com/joho/godotenv/autoload"

	"github.com/ConsenSys/fc-retrieval-provider/config"
	"github.com/ConsenSys/fc-retrieval-provider/pkg/provider"
)

// Start Provider service
func main() {
	conf := config.NewConfig()
	log.Init(conf)

	log.Info("Start Provider service...")
	p := provider.NewProvider(conf)
	p.Start()
}
