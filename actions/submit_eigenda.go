package actions

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Layr-Labs/eigenda/api/grpc/disperser"
	"github.com/Layr-Labs/eigenda/clients"
	"github.com/Layr-Labs/eigenda/core/auth"
	"github.com/Layr-Labs/eigenda/encoding/utils/codec"
	"github.com/stretchr/testify/assert"
)

type sendEigenDAAction struct {
	DisperserClient *clients.DisperserClient
	AuthKey         string
}

func NewSendEigenDAAction(authKey string) *sendEigenDAAction {
	config := clients.NewConfig(
		"disperser-holesky.eigenda.xyz",
		"443",
		time.Second*10,
		true,            
	)

	signer := auth.NewSigner(authKey)
	client := clients.NewDisperserClient(config, signer)

	return &sendEigenDAAction{
		DisperserClient: client,
		AuthKey:         authKey,
	}
}

func (e *sendEigenDAAction) Execute(ctx context.Context, data []byte) error {
	data = codec.ConvertByPaddingEmptyByte(data)
	quorums := []uint8{}
	blobStatus, requestID, err := e.DisperserClient.DisperseBlob(ctx, data, quorums)
	if err != nil || *blobStatus == disperser.Failed {
		return fmt.Errorf("error dispersing blob: %v", err)
	}

	fmt.Printf("Initial Blob Status: %+v\n", blobStatus)
	fmt.Printf("Request ID: %s\n", string(requestID))

	statusCtx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			statusReply, err := e.DisperserClient.GetBlobStatus(statusCtx, requestID)
			if err != nil {
				return fmt.Errorf("error getting blob status: %v", err)
			}

			if statusReply.Status == disperser_rpc.BlobStatus_FINALIZED {
				fmt.Printf("Blob Status is finalized: %s\n", pprint(statusReply))
				return nil
			} else if statusReply.Status == disperser_rpc.BlobStatus_FAILED {
				return fmt.Errorf("error dispersing blob: %v", statusReply.Status)
			} else {
				fmt.Printf("Current Blob Status: %s\n", pprint(statusReply))
			}
		case <-statusCtx.Done():
			return fmt.Errorf("timed out waiting for blob to finalize")
		}
	}
}

func pprint(m proto.Message) string {
	marshaler := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	jsonBytes, err := marshaler.Marshal(m)
	if err != nil {
		panic("Failed to marshal proto to JSON")
	}
	return string(jsonBytes)
}
