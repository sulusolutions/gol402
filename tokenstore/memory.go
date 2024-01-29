package tokenstore

import (
	"fmt"
	"net/url"
	"sync"
)

// InMemoryStore is an in-memory implementation of the TokenStore interface.
type InMemoryStore struct {
	mu    sync.RWMutex
	hosts map[string]*trie // Map[Host]*pathTrie
}

// NewInMemoryStore creates a new instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		hosts: make(map[string]*trie),
	}
}

// Put saves a token against a specified host and path from the URL.
func (ims *InMemoryStore) Put(u *url.URL, token *Token) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	host := u.Host
	path := u.Path

	// Detect invalid host
	if host == "" {
		return fmt.Errorf("invalid empty host.")
	}

	if ims.hosts[host] == nil {
		ims.hosts[host] = newTrie()
	}

	return ims.hosts[host].put(path, token)
}

// Get looks for a token that matches the given URL.
// It returns the most relevant token if available; otherwise, it returns nil.
func (ims *InMemoryStore) Get(u *url.URL) (*Token, bool) {
	ims.mu.RLock()
	defer ims.mu.RUnlock()

	host := u.Host
	path := u.Path

	if trie, ok := ims.hosts[host]; ok {
		return trie.get(path)
	}

	return nil, false
}

// Delete removes a token that matches the given URL.
func (ims *InMemoryStore) Delete(u *url.URL) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	host := u.Host
	path := u.Path

	if trie, ok := ims.hosts[host]; ok {
		return trie.delete(path)
	}

	return nil
}
