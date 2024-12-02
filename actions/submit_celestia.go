package actions

import (
	"bytes"
	"context"
	"fmt"

	client "github.com/celestiaorg/celestia-openrpc"
	"github.com/celestiaorg/celestia-openrpc/types/blob"
	"github.com/celestiaorg/celestia-openrpc/types/share"
	blobtypes "github.com/celestiaorg/celestia-app/x/blob/types"
)

type SendCelestiaAction struct {
	URL   string
	Token string
	Data  string
}

func (a *SendCelestiaAction) Execute(ctx context.Context) (int64, error) {
	client, err := client.NewClient(ctx, a.URL, a.Token)
	if err != nil {
		return 0, fmt.Errorf("failed to initialize Celestia client: %w", err)
	}
	namespace, err := share.NewBlobNamespaceV0([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	if err != nil {
		return 0, fmt.Errorf("failed to create namespace: %w", err)
	}
	dataBlob, err := blob.NewBlobV0(namespace, []byte(a.Data))
	if err != nil {
		return 0, fmt.Errorf("failed to create blob: %w", err)
	}
	sizeOfDataInBytes := len(a.Data)
	gasLimit := blobtypes.DefaultEstimateGas([]uint32{uint32(sizeOfDataInBytes)})
	height, err := client.Blob.Submit(ctx, []*blob.Blob{dataBlob}, gasLimit)
	if err != nil {
		return 0, fmt.Errorf("failed to submit blob: %w", err)
	}
	retrievedBlobs, err := client.Blob.GetAll(ctx, height, []share.Namespace{namespace})
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve blob: %w", err)
	}
	if !bytes.Equal(dataBlob.Commitment, retrievedBlobs[0].Commitment) {
		return 0, fmt.Errorf("retrieved blob does not match submitted blob")
	}
	fmt.Printf("Blob was included at height %d\n", height)
	return height, nil
}
