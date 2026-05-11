package cidr

import (
	"strings"
	"testing"
)

func TestExcludeFromLargePool(t *testing.T) {
	pool := NewPool()

	bigCIDR, err := Parse("10.0.0.0/8")
	if err != nil {
		t.Fatalf("failed to parse 10.0.0.0/8: %v", err)
	}
	pool.AddAvailable(bigCIDR)

	excludeCIDR, err := Parse("10.0.1.0/24")
	if err != nil {
		t.Fatalf("failed to parse 10.0.1.0/24: %v", err)
	}

	err = pool.Exclude(excludeCIDR)
	if err != nil {
		t.Fatalf("failed to exclude: %v", err)
	}

	t.Logf("Available CIDRs after exclusion (%d):", len(pool.Available))
	for _, c := range pool.Available {
		t.Logf("  %s", c.String())
	}

	t.Logf("Used CIDRs (%d):", len(pool.Used))
	for _, c := range pool.Used {
		t.Logf("  %s", c.String())
	}

	foundUsed := false
	for _, c := range pool.Used {
		if c.String() == "10.0.1.0/24" {
			foundUsed = true
			break
		}
	}
	if !foundUsed {
		t.Error("10.0.1.0/24 should be in used pool")
	}

	if len(pool.Available) == 0 {
		t.Error("available pool should not be empty")
	}

	if len(pool.Available) == 1 && pool.Available[0].String() == "10.0.0.0/24" {
		t.Error("BUG: only 10.0.0.0/24 remains, should have much more available space")
	}

	var totalAddresses uint64
	for _, c := range pool.Available {
		totalAddresses += c.Size()
	}
	totalAddresses += 256

	expectedTotal := uint64(1) << 24
	if totalAddresses != expectedTotal {
		t.Errorf("total addresses should be %d, got %d", expectedTotal, totalAddresses)
	}
}

func TestExcludeSimple(t *testing.T) {
	pool := NewPool()

	cidr16, err := Parse("192.168.0.0/16")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	pool.AddAvailable(cidr16)

	exclude, err := Parse("192.168.1.0/24")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	err = pool.Exclude(exclude)
	if err != nil {
		t.Fatalf("failed to exclude: %v", err)
	}

	t.Logf("Available after exclude (%d):", len(pool.Available))
	for _, c := range pool.Available {
		t.Logf("  %s", c.String())
	}

	var totalAvailable uint64
	for _, c := range pool.Available {
		totalAvailable += c.Size()
	}

	expectedAvailable := uint64(1)<<16 - uint64(1)<<8
	if totalAvailable != expectedAvailable {
		t.Errorf("expected %d available addresses, got %d", expectedAvailable, totalAvailable)
	}
}

func TestExcludeEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		available   string
		exclude     string
		expectedLen int
		checkAddr   string
		shouldExist bool
	}{
		{
			name:        "exclude left edge",
			available:   "10.0.0.0/16",
			exclude:     "10.0.0.0/24",
			expectedLen: 1,
			checkAddr:   "10.0.1.0/24",
			shouldExist: false,
		},
		{
			name:        "exclude right edge",
			available:   "10.0.0.0/16",
			exclude:     "10.0.255.0/24",
			expectedLen: 1,
			checkAddr:   "10.0.254.0/24",
			shouldExist: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool := NewPool()
			avail, _ := Parse(tc.available)
			exc, _ := Parse(tc.exclude)
			pool.AddAvailable(avail)
			err := pool.Exclude(exc)
			if err != nil {
				t.Fatalf("exclude failed: %v", err)
			}

			t.Logf("Available (%d):", len(pool.Available))
			for _, c := range pool.Available {
				t.Logf("  %s", c.String())
			}

			if len(pool.Available) < tc.expectedLen {
				t.Errorf("expected at least %d available CIDRs, got %d", tc.expectedLen, len(pool.Available))
			}

			if !tc.shouldExist {
				for _, c := range pool.Used {
					if strings.Contains(c.String(), tc.checkAddr[:strings.Index(tc.checkAddr, "/")]) {
						t.Logf("checking %s in used pool", tc.checkAddr)
					}
				}
			}
		})
	}
}
