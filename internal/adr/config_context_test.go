package adr

import (
	"context"
	"testing"
)

func TestLoadConfigContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := loadConfigContext(ctx, t.TempDir()); err == nil {
		t.Fatalf("expected canceled error")
	}
}

func TestSaveConfigContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := saveConfigContext(ctx, t.TempDir(), DefaultConfig()); err == nil {
		t.Fatalf("expected canceled error")
	}
}
