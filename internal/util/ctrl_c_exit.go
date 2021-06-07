/*
Package util - contains a collection of different functions used across all Retrieval Provider instances.

Utility function like ones to work with time, process lifetime and binary versioning.
*/
package util

// Copyright (C) 2020 ConsenSys Software Inc

import (
	"os"
	"os/signal"
)

// SetUpCtrlCExit configures the program such that when Control-C is hit, gracefulExit is called, followed by program exit.
func SetUpCtrlCExit(gracefulExit func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		gracefulExit()
		os.Exit(0)
	}()
}
