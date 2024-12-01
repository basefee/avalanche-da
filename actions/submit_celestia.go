package actions

import (
	"bytes"
	"context"
	"fmt"

	client "github.com/celestiaorg/celestia-openrpc"
	"github.com/celestiaorg/celestia-openrpc/types/blob"
	"github.com/celestiaorg/celestia-openrpc/types/share"
)

type SendCelestiaAction struct {
	URL       string 
	Token     string 
	Namespace []byte 
	Data      []byte 
}

func SendCelestiaBlob(ctx context.Context, url string, token string, namespace []byte, data []byte) (string, error) {
	celestiaClient, err := client.NewClient(ctx, url, token)
	if err != nil {
		return "", fmt.Errorf("failed to create Celestia client: %w", err)
	}
	blobNamespace, err := share.NewBlobNamespaceV0(namespace)
	if err != nil {
		return "", fmt.Errorf("failed to create namespace: %w", err)
	}
	newBlob, err := blob.NewBlobV0(blobNamespace, data)
	if err != nil {
		return "", fmt.Errorf("failed to create blob: %w", err)
	}
	height, err := celestiaClient.Blob.Submit(ctx, []*blob.Blob{newBlob}, blob.DefaultGasPrice())
	if err != nil {
		return "", fmt.Errorf("failed to submit blob: %w", err)
	}
	retrievedBlobs, err := celestiaClient.Blob.GetAll(ctx, height, []share.Namespace{blobNamespace})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve blobs: %w", err)
	}
	if !bytes.Equal(newBlob.Commitment, retrievedBlobs[0].Commitment) {
		return "", fmt.Errorf("blob verification failed: retrieved blob does not match submitted blob")
	}

	return fmt.Sprintf("Blob successfully included at height %d", height), nil
}

func (a *SendCelestiaAction) Execute(ctx context.Context) (string, error) {
	result, err := SendCelestiaBlob(ctx, a.URL, a.Token, a.Namespace, a.Data)
	if err != nil {
		return "", fmt.Errorf("failed to execute SendCelestiaAction: %w", err)
	}
	return result, nil
}