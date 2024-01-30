package tokenstore

import (
	"fmt"
	"net/url"
	"sync"
	"testing"
)

// TestPutNewToken verifies that a new token can be added successfully.
func TestPutNewToken(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	want := Token("token123")

	err := store.Put(testURL, want)
	if err != nil {
		t.Errorf("Failed to put new token: %v", err)
	}

	got, ok := store.Get(testURL)
	if !ok || got != want {
		t.Errorf("Expected token %v, got %v", want, got)
	}
}

// TestUpdateToken verifies that an existing token can be updated.
func TestUpdateToken(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	initialToken := Token("initialToken")
	updatedToken := Token("updatedToken")

	_ = store.Put(testURL, initialToken)

	err := store.Put(testURL, updatedToken)
	if err != nil {
		t.Errorf("Failed to update token: %v", err)
	}

	got, ok := store.Get(testURL)
	if !ok {
		t.Errorf("Token does not exist for URL: %v", testURL)
	}
	if got != updatedToken {
		t.Errorf("Expected updated token %v, got %v", updatedToken, got)
	}
}

// TestPutDifferentPaths verifies that tokens for the same host but different paths are stored separately.
func TestPutDifferentPaths(t *testing.T) {
	store := NewInMemoryStore()
	host := "http://host.com"

	tests := []struct {
		path  string
		token Token
	}{
		{"/path1", Token("token1")},
		{"/path2", Token("token2")},
	}

	// Put tokens
	for _, tc := range tests {
		testURL := &url.URL{Host: host, Path: tc.path}
		if err := store.Put(testURL, tc.token); err != nil {
			t.Fatalf("Failed to put token for %s: %v", tc.path, err)
		}
	}

	// Get and check tokens
	for _, tc := range tests {
		testURL := &url.URL{Host: host, Path: tc.path}
		got, ok := store.Get(testURL)
		if !ok {
			t.Errorf("Token for path %q not found", tc.path)
		} else if got != tc.token {
			t.Errorf("For path %q, expected token %q, got %q", tc.path, tc.token, got)
		}

	}
}

// TestConcurrentPut verifies that concurrent Put operations do not cause race conditions or data corruption.
func TestConcurrentPut(t *testing.T) {
	store := NewInMemoryStore()
	baseURL := "http://host.com/path"
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			testURL, err := url.Parse(baseURL + fmt.Sprint(i))
			if err != nil {
				t.Errorf("Failed to parse URL: %v", err)
			}
			if testURL == nil {
				t.Errorf("Failed to parse URL: %v", baseURL+fmt.Sprint(i))
			}
			want := Token("token" + fmt.Sprint(i))
			_ = store.Put(testURL, want)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 100; i++ {
		testURL, _ := url.Parse(baseURL + fmt.Sprint(i))
		got, ok := store.Get(testURL)
		want := Token("token" + fmt.Sprint(i))
		if !ok {
			t.Errorf("Concurrent put failed for %v: token not found", testURL)
		}
		if got != want {
			t.Errorf("Concurrent put failed for %v: expected %v, got %v", testURL, want, got)
		}
	}
}

// TestGetTokenExactMatch verifies retrieving a token for a URL that matches both host and path.
func TestGetTokenExactMatch(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	token := Token("token123")

	_ = store.Put(testURL, token)

	got, ok := store.Get(testURL)
	if !ok {
		t.Errorf("Expected to retrieve token %v, but retrieval failed", token)
	}
	if got != token {
		t.Errorf("Expected to retrieve token %v, got %v", token, got)
	}
}

// TestGetTokenHostMatch verifies retrieving a token for a URL that matches only the host.
func TestGetTokenHostMatch(t *testing.T) {
	store := NewInMemoryStore()
	putURL, _ := url.Parse("http://host.com/path")
	getURL, _ := url.Parse("http://host.com/anotherpath")
	want := Token("token123")

	_ = store.Put(putURL, want)

	got, ok := store.Get(getURL)
	if !ok || got != want {
		t.Errorf("Expected to retrieve token %v for host match, got %v", want, got)
	}
}

// TestGetTokenNoMatch verifies that no token is retrieved for a URL with no match.
func TestGetTokenNoMatch(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")

	_, ok := store.Get(testURL)
	if ok {
		t.Errorf("Expected no token to be retrieved, but got one")
	}
}

// TestConcurrentGet verifies that concurrent Get operations do not cause race conditions.
func TestConcurrentGet(t *testing.T) {
	store := NewInMemoryStore()
	u, err := url.Parse("http://host.com/path")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	want := Token("token123")
	_ = store.Put(u, want)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, ok := store.Get(u)
			if !ok {
				t.Errorf("Concurrent Get failed: token not found")
			}
			if got != want {
				t.Errorf("Concurrent Get failed: expected %v, got %v", want, got)
			}
		}()
	}

	wg.Wait()
}

// TestDeleteExistingToken verifies deleting an existing token.
func TestDeleteExistingToken(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	want := Token("token123")

	_ = store.Put(testURL, want)
	_ = store.Delete(testURL)

	got, ok := store.Get(testURL)
	if ok {
		t.Errorf("Expected no token after deletion, but found %v", got)
	}
}

// TestDeleteNonExistentToken ensures deleting a non-existent token does not cause errors.
func TestDeleteNonExistentToken(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")

	err := store.Delete(testURL)
	if err != nil {
		t.Errorf("Deleting non-existent token should not cause error: %v", err)
	}
}

// TestDeleteEffectOnOtherTokens verifies that deleting a token does not affect other tokens.
func TestDeleteEffectOnOtherTokens(t *testing.T) {
	store := NewInMemoryStore()
	host := "http://host.com"
	firstPath, secondPath := "/path1", "/path2"
	firstToken, secondToken := Token("token1"), Token("token2")

	// Store two tokens under the same host but different paths
	if err := store.Put(&url.URL{Host: host, Path: firstPath}, firstToken); err != nil {
		t.Fatalf("Failed to store token for first path: %v", err)
	}
	if err := store.Put(&url.URL{Host: host, Path: secondPath}, secondToken); err != nil {
		t.Fatalf("Failed to store token for second path: %v", err)
	}

	// Delete the token associated with the first path
	if err := store.Delete(&url.URL{Host: host, Path: firstPath}); err != nil {
		t.Fatalf("Failed to delete token for first path: %v", err)
	}

	// Attempt to retrieve the deleted token
	got, ok := store.Get(&url.URL{Host: host, Path: firstPath})
	if !ok {
		t.Error("Expected token2 for the first path after deletion")
	}
	if got != secondToken {
		t.Errorf("Expected token2 for the first path after deletion, got %v", got)
	}

	// Verify the second token is unaffected
	got, ok = store.Get(&url.URL{Host: host, Path: secondPath})
	if !ok {
		t.Error("Expected to find a token for the second path, but none was found")
	} else if got != secondToken {
		t.Errorf("Expected to retrieve the second token %q, but got %q", secondToken, got)
	}
}

// TestDeleteHostLevelCleanup verifies that deleting the last token for a host removes the host entry.
func TestDeleteHostLevelCleanup(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	want := Token("token123")

	_ = store.Put(testURL, want)
	_ = store.Delete(testURL)

	// Attempt to retrieve a token for the same host but different path
	otherPathURL, _ := url.Parse("http://host.com/otherpath")
	_, ok := store.Get(otherPathURL)

	if ok {
		t.Errorf("Host entry should be removed after deleting its last token")
	}
}

// TestConcurrentDelete verifies that concurrent Delete operations do not cause race conditions.
func TestConcurrentDelete(t *testing.T) {
	store := NewInMemoryStore()
	baseURL, err := url.Parse("http://host.com/path")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	want := Token("token123")
	_ = store.Put(baseURL, want)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Delete(&url.URL{Host: "host.com", Path: "/path"})
		}()
	}

	wg.Wait()

	_, ok := store.Get(&url.URL{Host: "host.com", Path: "/path"})
	if ok {
		t.Errorf("Token should have been deleted after concurrent deletion attempts")
	}
}
