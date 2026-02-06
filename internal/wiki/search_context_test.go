package wiki

import (
	"context"
	"testing"
)

func TestSearchPagesContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := SearchPagesContext(ctx, t.TempDir(), SearchOptions{Query: "hello"}); err == nil {
		t.Fatalf("expected canceled error")
	}
}
