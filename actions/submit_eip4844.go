package actions

import (
	"context"
	"fmt"
	"math/big"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

const BlobTxDataLimit = 128 * 1024 // Example limit for blob data (128 KB)

type SendBlobAction struct {
	Client     *ethclient.Client
	PrivateKey string
	Data       []byte 
}

type SendBlobActionResult struct {
    TransactionHash string `json:"transaction_hash"`
    Status          string `json:"status"`
    Units           uint64 `json:"units"`
    Error           string `json:"error,omitempty"`
}

func (s *SendBlobAction) GetTypeID() uint8 {
    return 1
}

func (s *SendBlobActionResult) GetTypeID() uint8 {
    return 2
}

func (a *SendBlobAction) ComputeUnits() (uint64, error) {

	return uint64(len(a.Data) / 1024), nil 
}

func (a *SendBlobAction) Execute(ctx context.Context, rawBlob []byte) (*SendBlobActionResult, error) {

    a.Data = rawBlob
    units, err := a.ComputeUnits()
    if err != nil {
        return nil, err
    }
    logrus.Infof("Computed units: %d", units)
    privateKey, err := crypto.HexToECDSA(a.PrivateKey)
    if err != nil {
        return nil, err
    }
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("error casting public key to ECDSA")
    }
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    nonce, err := a.Client.PendingNonceAt(ctx, fromAddress)
    if err != nil {
        return nil, err
    }
    gasTipCap, err := a.Client.SuggestGasTipCap(ctx)
    if err != nil {
        return nil, err
    }

    gasFeeCap, err := a.Client.SuggestGasPrice(ctx)
    if err != nil {
        return nil, err
    }

    chainID, err := a.Client.ChainID(ctx)
    if err != nil {
        return nil, err
    }

    blobData := make([]byte, BlobTxDataLimit)
    copy(blobData, rawBlob)

    blob := kzg4844.Blob(blobData)

    blobs := []kzg4844.Blob{blob}
    sideCar := makeSidecar(blobs)

    blobHashes := sideCar.BlobHashes()

    blobFeeCap := new(big.Int)
    blobFeeCap.SetString("1000000000", 10) // 1 Gwei as an example

    tx := types.NewTx(&types.BlobTx{
        ChainID:   uint256.MustFromBig(chainID),
        Nonce:     uint64(nonce),
        GasTipCap: uint256.MustFromBig(gasTipCap),
        GasFeeCap: uint256.MustFromBig(gasFeeCap),
        Gas:       uint64(21000), // Example gas limit
        To:        common.HexToAddress("0xb10000000000000000000000000000000000000b"),
        BlobFeeCap: uint256.MustFromBig(blobFeeCap),
        BlobHashes: blobHashes,
        Sidecar:    sideCar,
    })

    auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
    if err != nil {
        return nil, err
    }

    signedTx, err := auth.Signer(auth.From, tx)
    if err != nil {
        return nil, err
    }

    err = a.Client.SendTransaction(ctx, signedTx)
    if err != nil {
        return nil, err
    }

    logrus.Infof("Sent blob transaction with hash: %s", signedTx.Hash().Hex())

    return &SendBlobActionResult{
        TransactionHash: signedTx.Hash().Hex(),
        Status:          "success",
        Units:           units,
    }, nil
}



func makeSidecar(blobs []kzg4844.Blob) *types.BlobTxSidecar {
	var (
		commitments []kzg4844.Commitment
		proofs      []kzg4844.Proof
	)

	for _, blob := range blobs {
		c, _ := kzg4844.BlobToCommitment(blob)
		p, _ := kzg4844.ComputeBlobProof(blob, c)

		commitments = append(commitments, c)
		proofs = append(proofs, p)
	}

	return &types.BlobTxSidecar{
		Blobs:       blobs,
		Commitments: commitments,
		Proofs:      proofs,
	}
}
