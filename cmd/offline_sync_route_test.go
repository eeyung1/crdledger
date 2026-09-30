package main

import (
	"os"
	"strings"
	"testing"
)

func TestOfflineTransactionSyncEndpointMatchesRegisteredRoute(t *testing.T) {
	client, err := os.ReadFile("../static/js/offline-transactions.js")
	if err != nil { t.Fatal(err) }
	mainSource, err := os.ReadFile("main.go")
	if err != nil { t.Fatal(err) }
	const endpoint = "/api/sync/transactions"
	if !strings.Contains(string(client), "fetch('"+endpoint+"'") { t.Fatalf("offline client must POST to %s", endpoint) }
	if !strings.Contains(string(mainSource), "mux.HandleFunc(\""+endpoint+"\"") { t.Fatalf("server must register %s", endpoint) }
}
