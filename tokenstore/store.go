package tokenstore

import "net/url"

// Token represents a wrapper around the L402 token string.
type Token string

// Store defines the interface for storing and retrieving L402 tokens.
type Store interface {
	// StoreToken saves a token against a specified host and path.
	Put(u *url.URL, token Token) error

	// RetrieveToken looks for a token that matches the given host and path.
	// It returns the most relevant token if available; otherwise, it returns nil.
	Get(u *url.URL) (Token, bool)

	// Delete removes a token that matches the given host and path.
	Delete(u *url.URL) error
}
