package tokenstore

import (
	"net/url"
	"sync"
)

type InMemoryStore struct {
	mu    sync.RWMutex
	store map[string]map[string]Token // Outer map key is host, inner map key is path
}

// NewInMemoryStore creates a new instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store: make(map[string]map[string]Token),
	}
}

// Put saves a token against a specified host and path from the URL.
func (ims *InMemoryStore) Put(u *url.URL, token Token) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	host := u.Host
	path := u.Path

	// Initialize host map if not present
	if _, exists := ims.store[host]; !exists {
		ims.store[host] = make(map[string]Token)
	}

	// Save token against host and path
	ims.store[host][path] = token

	return nil
}

// Get looks for a token that matches the given URL.
// It returns the most relevant token if available; otherwise, it returns nil.
func (ims *InMemoryStore) Get(u *url.URL) (Token, bool) {
	ims.mu.RLock()
	defer ims.mu.RUnlock()

	host := u.Host
	path := u.Path

	// Check if host exists
	if paths, hostExists := ims.store[host]; hostExists {
		// Attempt to get the exact path match first
		if token, pathExists := paths[path]; pathExists {
			return token, true
		}

		// If no exact path match, take a token from any path under the host
		for _, token := range paths {
			return token, true
		}
	}

	return "", false
}

// Delete removes a token that matches the given URL.
func (ims *InMemoryStore) Delete(u *url.URL) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	host := u.Host
	path := u.Path

	// Check if host exists
	if paths, hostExists := ims.store[host]; hostExists {
		// Remove the path entry
		delete(paths, path)

		// If the inner map is now empty, remove the host entry as well
		if len(paths) == 0 {
			delete(ims.store, host)
		}
	}

	return nil
}
