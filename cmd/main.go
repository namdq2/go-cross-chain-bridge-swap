package main

import (
	"log"
	"os"
	"strings"

	"github.com/namdq2/go-cross-chain-bridge-swap/internal/api"
	"github.com/namdq2/go-cross-chain-bridge-swap/internal/models"
	"github.com/namdq2/go-cross-chain-bridge-swap/internal/service"
)

func main() {
	// Initialize database
	db, err := models.NewDatabase(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize bridge service
	privateKeys := strings.Split(os.Getenv("HOT_WALLET_PRIVATE_KEYS"), ",")
	bridgeService, err := service.NewBridgeService(service.Config{
		Chain1RPC:   os.Getenv("CHAIN1_RPC"),
		Chain2RPC:   os.Getenv("CHAIN2_RPC"),
		Chain1ID:    1,  // Ethereum Mainnet
		Chain2ID:    56, // BSC Mainnet
		BridgeAddr1: os.Getenv("BRIDGE_ADDR1"),
		BridgeAddr2: os.Getenv("BRIDGE_ADDR2"),
	}, privateKeys, db)
	if err != nil {
		log.Fatalf("Failed to initialize bridge service: %v", err)
	}

	// Start API server
	server := api.NewServer(bridgeService)
	log.Fatal(server.Start(":8080"))
}
