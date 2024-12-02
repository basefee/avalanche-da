package actions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ava-labs/hypersdk-starter-kit/actions"
)

func TestSendAvailAction(t *testing.T) {
	seed := "SEED_PHRASE"
	data := "Unfold24 CoinDCX"
	networkURL := "AVAIL_URL"

	action := actions.SendAvailAction{
		Seed:       seed,
		Data:       data,
		NetworkURL: networkURL,
	}

	ctx := context.Background()
	blockHash, txHash, err := action.Execute(ctx)

	assert.NoError(t, err, "Expected no error from Execute")
	assert.NotEmpty(t, blockHash, "Block hash should not be empty")
	assert.NotEmpty(t, txHash, "Transaction hash should not be empty")
}
