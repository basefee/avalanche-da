package tests

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/ava-labs/hypersdk-starter-kit/actions"
)

func TestSendEigenDAAction(t *testing.T) {
	authKey := "EIGENDA_AUTH_PK"
	sendAction := actions.NewSendEigenDAAction(authKey)

	rawData := []byte("example data to disperse")

	err := sendAction.Execute(context.Background(), rawData)
	if err != nil {
		t.Fatalf("Failed to send data to EigenDA: %v", err)
	}

	assert.NoError(t, err, "SendEigenDAAction should not return an error")
}
