package tokenstore

import (
	"net/url"
	"testing"
)

func TestInMemoryStore_Put(t *testing.T) {
	s := NewInMemoryStore()

	// Create a sample URL
	u, err := url.Parse("https://example.com/path/to/resource")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	token := Token("sampleToken")

	// Test Put
	err = s.Put(u, &token)
	if err != nil {
		t.Errorf("Put failed with error: %v", err)
	}

	// Retrieve the token and check if it matches
	retrievedToken, found := s.Get(u)
	if !found {
		t.Errorf("Token not found for URL: %s", u.String())
	} else if *retrievedToken != token {
		t.Errorf("Expected token %s, but got %s", token, *retrievedToken)
	}
}

func TestInMemoryStore_PutError(t *testing.T) {
	s := NewInMemoryStore()
	token := Token("sampleToken")

	// Test Put with an invalid URL
	err := s.Put(&url.URL{}, &token)
	if err == nil {
		t.Errorf("Expected error when putting with invalid URL, but got nil")
	}
}

func TestInMemoryStore_GetErrorMissingHost(t *testing.T) {
	// Create a new InMemoryStore
	store := NewInMemoryStore()

	// Create a sample URL with no host
	u, err := url.Parse("missing.com/path/to/resource")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	// Test Get with a URL missing the host
	token, found := store.Get(u)
	if found {
		t.Errorf("Expected Get to return not found, but it found a token: %s", *token)
	}
}
