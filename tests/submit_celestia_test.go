package actions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ava-labs/hypersdk-starter-kit/actions"
)

func TestSendCelestiaAction(t *testing.T) {
	action := actions.SendCelestiaAction{
		URL:       "TESTNET_NODE_URL", // You need to run a celestia local node
		Token:     "TEST_TOKEN",                     // Replace with a test token
		Namespace: []byte{0xDE, 0xAD, 0xBE, 0xEF},
		Data:      []byte("Test Blob Data"),
	}

	ctx := context.Background()
	result, err := action.Execute(ctx)
	assert.NoError(t, err, "Expected no error from Execute")
	assert.Contains(t, result, "Blob successfully included at height", "Expected success message in result")
}
