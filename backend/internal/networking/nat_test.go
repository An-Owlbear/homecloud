package networking

import (
	"context"
	"flag"
	"testing"

	"github.com/An-Owlbear/homecloud/backend/internal/config"
)

var nat = flag.Bool("nat", false, "Whether or not to run nat tests")

// IMPORTANT - this test will likely be inconsistent depending on your network
// If your network doesn't support uPnP or NAT-PMP this won't work
func TestTryMapPort(t *testing.T) {
	if !*nat {
		t.Skip("Skipping tests when -nat is disabled")
	}

	err := TryMapPort(context.Background(), 13231, 1323, config.NewDeviceConfig())
	if err != nil {
		t.Errorf("TestTryMapPort failed: %v", err)
	}
}
