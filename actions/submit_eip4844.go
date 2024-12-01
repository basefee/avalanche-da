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
	"github.com/ethereum/go-ethereum/crypto/kzg4844" // Assuming kzg4844 is imported here
)

const BlobTxDataLimit = 128 * 1024 // Example limit for blob data (128 KB)

type SendBlobAction struct {
	Client     *ethclient.Client
	PrivateKey string
}

func (a *SendBlobAction) Execute(ctx context.Context, rawBlob []byte) error {
	// Load private key
	privateKey, err := crypto.HexToECDSA(a.PrivateKey)
	if err != nil {
		return err
	}

	// Extract public address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Fetch current nonce
	nonce, err := a.Client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	// Fetch gas and chain ID
	gasTipCap, err := a.Client.SuggestGasTipCap(ctx)
	if err != nil {
		return err
	}

	gasFeeCap, err := a.Client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	chainID, err := a.Client.ChainID(ctx)
	if err != nil {
		return err
	}

	// Create blob data (ensure proper size limit)
	blobData := make([]byte, BlobTxDataLimit)
	copy(blobData, rawBlob)

	// Create kzg4844.Blob from rawBlob (assuming a constructor or conversion exists)
	// Here, we assume kzg4844.Blob can be constructed with a byte slice, replace as necessary
	blob := kzg4844.Blob(blobData)

	// Create sidecar and hashes using makeSidecar function
	blobs := []kzg4844.Blob{blob} // Passing the converted blob to the sidecar
	sideCar := makeSidecar(blobs)

	blobHashes := sideCar.BlobHashes()

	// Construct the transaction
	blobFeeCap := new(big.Int)
	blobFeeCap.SetString("1000000000", 10) // Example fee cap value

	tx := types.NewTx(&types.BlobTx{
		ChainID:   uint256.MustFromBig(chainID),
		Nonce:     uint64(nonce),
		GasTipCap: uint256.MustFromBig(gasTipCap),
		GasFeeCap: uint256.MustFromBig(gasFeeCap),
		Gas:       uint64(21000), // Minimum gas for a transaction
		To:        common.HexToAddress("0xb10000000000000000000000000000000000000b"),
		BlobFeeCap: uint256.MustFromBig(blobFeeCap),
		BlobHashes: blobHashes,
		Sidecar:    sideCar,
	})

	// Sign the transaction
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return err
	}

	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return err
	}

	// Send the transaction
	err = a.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	logrus.Infof("Sent blob transaction with hash: %s", signedTx.Hash().Hex())
	return nil
}

// makeSidecar function as defined earlier
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
