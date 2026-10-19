package main

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func TestTestFrameworkPreRunEReturnsInitializationError(t *testing.T) {
	t.Setenv("TEST_PROVIDER", `{"type":"aws"}`)

	initializationError := errors.New("probe failed")
	command := &cobra.Command{}
	command.SetContext(context.Background())
	preRunE := newTestFrameworkPreRunE(func(_ context.Context, provider string) error {
		if provider != `{"type":"aws"}` {
			t.Fatalf("expected TEST_PROVIDER value, got %q", provider)
		}
		return initializationError
	})

	err := preRunE(command, nil)
	if !errors.Is(err, initializationError) {
		t.Fatalf("expected initialization error to be returned, got %v", err)
	}
}
