package tests

import (
	"context"
	"testing"
	"github.com/ava-labs/hypersdk-starter-kit/actions"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TestSendBlobAction(t *testing.T) {
	client, err := ethclient.Dial("RPC") 
	if err != nil {
		t.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	sendAction := actions.SendBlobAction{
		Client:     client,
		PrivateKey: "ca834b039044a6fcd17f314dbd12e8bf3afc904e94e6200e1b5580b0eab4847c", // Replace with actual private key for testing
	}
	rawBlob := []byte("example raw blob data")
	result, err := sendAction.Execute(context.Background(), rawBlob)
	if err != nil {
		t.Fatalf("Failed to send blob: %v", err)
	}

	assert.Equal(t, "success", result.Status, "Expected status to be 'Success'")
	assert.NotEmpty(t, result.TransactionHash, "Expected transaction hash to be non-empty")
	assert.Equal(t, uint64(len(rawBlob)/1024), result.Units, "Expected the correct unit calculation")
}
