package actions

import (
	"context"
	"fmt"

	"github.com/availproject/avail-go-sdk/src/config"
	"github.com/availproject/avail-go-sdk/src/sdk"
	"github.com/availproject/avail-go-sdk/src/sdk/tx"
)

type SendAvailAction struct {
	Seed       string
	Data       string
	NetworkURL string
}

func (a *SendAvailAction) Execute(ctx context.Context) (string, string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}

	api, err := sdk.NewSDK(a.NetworkURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize Avail SDK: %w", err)
	}

	blockHash, txHash, err := tx.SubmitData(api, cfg.Seed, 1, a.Data, sdk.BlockInclusion)
	if err != nil {
		return "", "", fmt.Errorf("failed to submit data to Avail DA: %w", err)
	}

	return blockHash.Hex(), txHash.Hex(), nil
}
