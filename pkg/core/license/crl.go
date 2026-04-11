package license

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CRLChecker performs optional Certificate Revocation List checks.
// Only active when DBX_LICENSE_CRL_URL is configured (enterprise only).
// If not configured, all IsRevoked calls return false (pure offline mode).
type CRLChecker struct {
	url      string
	cacheTTL time.Duration
	client   *http.Client

	mu          sync.RWMutex
	revokedJTIs map[string]bool
	lastFetch   time.Time
}

// crlResponse is the decoded CRL payload (simplified; real CRL is a signed JWT).
type crlResponse struct {
	RevokedJTIs []string `json:"revoked_jtis"`
}

// NewCRLChecker creates a new CRL checker.
// If url is empty, the checker is a no-op (pure offline mode per spec 14.2).
func NewCRLChecker(url string, cacheTTL time.Duration) *CRLChecker {
	return &CRLChecker{
		url:         url,
		cacheTTL:    cacheTTL,
		client:      &http.Client{Timeout: 10 * time.Second},
		revokedJTIs: make(map[string]bool),
	}
}

// IsRevoked checks if a JTI is in the revocation list.
// Returns false if CRL is not configured or cache is empty (offline-first).
func (c *CRLChecker) IsRevoked(jti string) bool {
	if c.url == "" {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.revokedJTIs[jti]
}

// Refresh fetches the CRL from the configured URL.
// Non-blocking: network failure is non-fatal (logged, not returned to caller in prod).
// Respects cache TTL: does not re-fetch if cache is still fresh.
func (c *CRLChecker) Refresh() error {
	if c.url == "" {
		return nil
	}

	c.mu.RLock()
	if !c.lastFetch.IsZero() && time.Since(c.lastFetch) < c.cacheTTL {
		c.mu.RUnlock()
		return nil // cache is fresh
	}
	c.mu.RUnlock()

	resp, err := c.client.Get(c.url)
	if err != nil {
		return fmt.Errorf("CRL fetch failed (proceeding offline): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CRL returned HTTP %d", resp.StatusCode)
	}

	var crl crlResponse
	if err := json.NewDecoder(resp.Body).Decode(&crl); err != nil {
		return fmt.Errorf("CRL decode failed: %w", err)
	}

	c.mu.Lock()
	c.revokedJTIs = make(map[string]bool, len(crl.RevokedJTIs))
	for _, jti := range crl.RevokedJTIs {
		c.revokedJTIs[jti] = true
	}
	c.lastFetch = time.Now()
	c.mu.Unlock()

	return nil
}

// IsConfigured returns true if a CRL URL is set.
func (c *CRLChecker) IsConfigured() bool {
	return c.url != ""
}
