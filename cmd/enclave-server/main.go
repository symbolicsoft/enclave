// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/symbolicsoft/enclave/v2/internal/server"
	"github.com/symbolicsoft/enclave/v2/internal/store"
	"github.com/symbolicsoft/enclave/v2/internal/version"
)

func main() {
	fmt.Println("enclave-server", version.VERSION_SERVER)
	handleSigInterrupt()
	server := server.EnclaveServer{
		CertFilePath:  filepath.Join("/", "home", "nadim", "certs", "enclave.sh.crt"),
		KeyFilePath:   filepath.Join("/", "home", "nadim", "certs", "enclave.sh.key"),
		ListenAddress: "172.233.242.130",
		ListenPort:    7070,
	}
	server.Start()
}

func handleSigInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("closing database...")
			store.CloseDatabase()
			os.Exit(0)
		}
	}()
}
