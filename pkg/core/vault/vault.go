// Package vault provides credential resolution via HashiCorp Vault with caching.
package vault

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Credential holds database credentials resolved from Vault or inline config.
type Credential struct {
	Username      string
	Password      string
	ConnectString string
	LeaseTTL      time.Duration
}

// Redacted returns a safe-to-log representation.
func (c *Credential) Redacted() string {
	if c.ConnectString != "" {
		return fmt.Sprintf("%s/***@%s", c.Username, c.ConnectString)
	}
	return fmt.Sprintf("%s/***", c.Username)
}

// Fetcher is a function that fetches credentials from Vault.
type Fetcher func(ctx context.Context, path string) (*Credential, error)

// Client manages credential resolution with caching.
type Client struct {
	fetcher  Fetcher
	cacheTTL time.Duration
	mu       sync.RWMutex
	cache    map[string]*cacheEntry
}

type cacheEntry struct {
	cred      *Credential
	fetchedAt time.Time
}

// Option configures the Client.
type Option func(*Client)

// WithFetcher sets a custom credential fetcher.
func WithFetcher(f Fetcher) Option {
	return func(c *Client) { c.fetcher = f }
}

// WithCacheTTL sets the credential cache duration.
func WithCacheTTL(d time.Duration) Option {
	return func(c *Client) { c.cacheTTL = d }
}

// NewClient creates a Vault client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		cacheTTL: 5 * time.Minute,
		cache:    make(map[string]*cacheEntry),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// GetCredential fetches a credential from Vault (or cache).
func (c *Client) GetCredential(ctx context.Context, path string) (*Credential, error) {
	if c.fetcher == nil {
		return nil, fmt.Errorf("vault not configured")
	}

	c.mu.RLock()
	if e, ok := c.cache[path]; ok && time.Since(e.fetchedAt) < c.cacheTTL {
		c.mu.RUnlock()
		return e.cred, nil
	}
	c.mu.RUnlock()

	cred, err := c.fetcher(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("vault fetch %s: %w", path, err)
	}

	c.mu.Lock()
	c.cache[path] = &cacheEntry{cred: cred, fetchedAt: time.Now()}
	c.mu.Unlock()

	return cred, nil
}

// ResolveCredential returns Vault creds or falls back to inline.
func (c *Client) ResolveCredential(ctx context.Context, credMode, vaultPath string, fallback *Credential) *Credential {
	if credMode == "vault" && vaultPath != "" && c.fetcher != nil {
		cred, err := c.GetCredential(ctx, vaultPath)
		if err == nil {
			return cred
		}
	}
	return fallback
}
