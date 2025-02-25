package networking

import (
	"context"
	"testing"
)

// IMPORTANT - this test will likely be inconsistent depending on your network
// If your network doesn't support uPnP or NAT-PMP this won't work
func TestTryMapPort(t *testing.T) {
	err := TryMapPort(context.Background(), 13231, 1323)
	if err != nil {
		t.Errorf("TestTryMapPort failed: %v", err)
	}
}
