package tests

import (
	"context"
	"testing"
	"github.com/ava-labs/hypersdk-starter-kit/actions"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TestSendBlobAction(t *testing.T) {
	// Assuming you have a way to create or mock the ethclient
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/iNDJZcDjmJIs-mOqC2t8UCskh0dThqyo")
	if err != nil {
		t.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Initialize SendBlobAction
	sendAction := actions.SendBlobAction{
		Client:     client,
		PrivateKey: "ca834b039044a6fcd17f314dbd12e8bf3afc904e94e6200e1b5580b0eab4847c", // Replace with actual private key
	}

	// Example blob data
	rawBlob := []byte("example raw blob data")

	// Execute the action
	err = sendAction.Execute(context.Background(), rawBlob)
	if err != nil {
		t.Fatalf("Failed to send blob: %v", err)
	}

	// Add assertions as needed for your test
	assert.NoError(t, err, "SendBlobAction should not return an error")
}
