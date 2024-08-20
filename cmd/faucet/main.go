// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/time/rate"

	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/examples/morpheusvm/actions"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"

	lconsts "github.com/ava-labs/hypersdk/examples/morpheusvm/consts"
	lrpc "github.com/ava-labs/hypersdk/examples/morpheusvm/rpc"
)

const amtStr = "10.00"

var (
	priv    ed25519.PrivateKey
	factory chain.AuthFactory
	cli     *rpc.JSONRPCClient
	lcli    *lrpc.JSONRPCClient
)

func init() {
	privBytes, err := hex.DecodeString(os.Getenv("FAUCET_PRIVATE_KEY_HEX"))
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}
	priv = ed25519.PrivateKey(privBytes)
	factory = auth.NewED25519Factory(priv)

	rpcEndpoint := os.Getenv("RPC_ENDPOINT")
	if rpcEndpoint == "" {
		rpcEndpoint = "http://localhost:9650"
	}
	url := fmt.Sprintf("%s/ext/bc/morpheusvm", rpcEndpoint)
	cli = rpc.NewJSONRPCClient(url)

	networkID, subnetID, chainID, err := cli.Network(context.TODO())
	if err != nil {
		log.Fatalf("failed to get network info: %v", err)
	}
	fmt.Println(networkID, subnetID, chainID)

	lcli = lrpc.NewJSONRPCClient(url, networkID, chainID)
}

func transferCoins(to string) (string, error) {
	toAddr, err := codec.ParseAddressBech32(lconsts.HRP, to)
	if err != nil {
		return "", fmt.Errorf("failed to parse to address: %w", err)
	}

	amt, err := utils.ParseBalance(amtStr, lconsts.Decimals)
	if err != nil {
		return "", fmt.Errorf("failed to parse amount: %w", err)
	}

	balanceBefore, err := lcli.Balance(context.TODO(), to)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}
	fmt.Printf("Balance before: %s\n", utils.FormatBalance(balanceBefore, lconsts.Decimals))

	// Check if balance is greater than 1.000
	threshold, _ := utils.ParseBalance("1.000", lconsts.Decimals)
	if balanceBefore > threshold {
		fmt.Printf("Balance is already greater than 1.000, no transfer needed\n")
		return "Balance is already greater than 1.000, no transfer needed", nil
	}

	parser, err := lcli.Parser(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to get parser: %w", err)
	}

	submit, _, _, err := cli.GenerateTransaction(
		context.TODO(),
		parser,
		[]chain.Action{&actions.Transfer{
			To:    toAddr,
			Value: amt,
		}},
		factory,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate transaction: %w", err)
	}

	err = submit(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}

	err = lcli.WaitForBalance(context.TODO(), to, amt)
	if err != nil {
		return "", fmt.Errorf("failed to wait for balance: %w", err)
	}

	return "Coins transferred successfully", nil
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/faucet/{address}", handleFaucetRequest).Methods("GET", "POST")

	// Create a new CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Wrap the router with the CORS handler
	handler := c.Handler(r)

	fmt.Println("Starting faucet server on port 8765\nOpen http://localhost:8765/faucet/morpheus1qqgvs58cq6f0fv876f2lccay8t55fwf6vg4c77h5c3h4gjruqelk5srn9ds to test the transfer")

	srv := &http.Server{
		Addr:         ":8765",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

var (
	ipLimiters = make(map[string]*rate.Limiter)
	mu         sync.Mutex
)

func handleFaucetRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for all responses
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get client IP
	clientIP := r.RemoteAddr

	// Check rate limit
	if !getRateLimiter(clientIP).Allow() {
		http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
		return
	}

	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		http.Error(w, "Address not provided", http.StatusBadRequest)
		return
	}

	message, err := transferCoins(address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to transfer coins: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, message)
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := ipLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(15*time.Second), 10)
		ipLimiters[ip] = limiter
	}

	return limiter
}