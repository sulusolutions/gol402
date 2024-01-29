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
	testToken := Token("token123")

	err := store.Put(testURL, &testToken)
	if err != nil {
		t.Errorf("Failed to put new token: %v", err)
	}

	storedToken, exists := store.Get(testURL)
	if !exists || *storedToken != testToken {
		t.Errorf("Expected token %v, got %v", testToken, storedToken)
	}
}

// TestUpdateToken verifies that an existing token can be updated.
func TestUpdateToken(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	initialToken := Token("initialToken")
	updatedToken := Token("updatedToken")

	_ = store.Put(testURL, &initialToken)
	err := store.Put(testURL, &updatedToken)
	if err != nil {
		t.Errorf("Failed to update token: %v", err)
	}

	storedToken, exists := store.Get(testURL)
	if !exists || *storedToken != updatedToken {
		t.Errorf("Expected updated token %v, got %v", updatedToken, storedToken)
	}
}

// TestPutDifferentPaths verifies that tokens for the same host but different paths are stored separately.
func TestPutDifferentPaths(t *testing.T) {
	store := NewInMemoryStore()
	host := "http://host.com"
	path1, path2 := "/path1", "/path2"
	token1, token2 := Token("token1"), Token("token2")

	_ = store.Put(&url.URL{Host: host, Path: path1}, &token1)
	_ = store.Put(&url.URL{Host: host, Path: path2}, &token2)

	storedToken1, exists1 := store.Get(&url.URL{Host: host, Path: path1})
	storedToken2, exists2 := store.Get(&url.URL{Host: host, Path: path2})

	if !exists1 || *storedToken1 != token1 {
		t.Errorf("Expected token %v for path %v, got %v", token1, path1, storedToken1)
	}
	if !exists2 || *storedToken2 != token2 {
		t.Errorf("Expected token %v for path %v, got %v", token2, path2, storedToken2)
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
			testToken := Token("token" + fmt.Sprint(i))
			_ = store.Put(testURL, &testToken)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 100; i++ {
		testURL, _ := url.Parse(baseURL + fmt.Sprint(i))
		storedToken, exists := store.Get(testURL)
		expectedToken := Token("token" + fmt.Sprint(i))
		if !exists || *storedToken != expectedToken {
			t.Errorf("Concurrent put failed for %v: expected %v, got %v", testURL, expectedToken, storedToken)
		}
	}
}

// TestGetTokenExactMatch verifies retrieving a token for a URL that matches both host and path.
func TestGetTokenExactMatch(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	testToken := Token("token123")

	_ = store.Put(testURL, &testToken)

	storedToken, exists := store.Get(testURL)
	if !exists || *storedToken != testToken {
		t.Errorf("Expected to retrieve token %v, got %v", testToken, storedToken)
	}
}

// TestGetTokenHostMatch verifies retrieving a token for a URL that matches only the host.
func TestGetTokenHostMatch(t *testing.T) {
	store := NewInMemoryStore()
	putURL, _ := url.Parse("http://host.com/path")
	getURL, _ := url.Parse("http://host.com/anotherpath")
	testToken := Token("token123")

	_ = store.Put(putURL, &testToken)

	storedToken, exists := store.Get(getURL)
	if !exists || *storedToken != testToken {
		t.Errorf("Expected to retrieve token %v for host match, got %v", testToken, storedToken)
	}
}

// TestGetTokenNoMatch verifies that no token is retrieved for a URL with no match.
func TestGetTokenNoMatch(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")

	_, exists := store.Get(testURL)
	if exists {
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
	testToken := Token("token123")
	_ = store.Put(u, &testToken)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			storedToken, exists := store.Get(u)
			if !exists || *storedToken != testToken {
				t.Errorf("Concurrent Get failed: expected %v, got %v", testToken, storedToken)
			}
		}()
	}

	wg.Wait()
}

// TestDeleteExistingToken verifies deleting an existing token.
func TestDeleteExistingToken(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	testToken := Token("token123")

	_ = store.Put(testURL, &testToken)
	_ = store.Delete(testURL)

	storedToken, exists := store.Get(testURL)
	if exists {
		t.Errorf("Expected no token after deletion, but found %v", storedToken)
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
	path1, path2 := "/path1", "/path2"
	token1, token2 := Token("token1"), Token("token2")

	_ = store.Put(&url.URL{Host: host, Path: path1}, &token1)
	_ = store.Put(&url.URL{Host: host, Path: path2}, &token2)

	_ = store.Delete(&url.URL{Host: host, Path: path1})

	storedToken1, _ := store.Get(&url.URL{Host: host, Path: path1})
	storedToken2, exists2 := store.Get(&url.URL{Host: host, Path: path2})

	if storedToken1 != storedToken2 {
		t.Errorf("Token for path1 should have been deleted and we should get token for path2")
	}
	if !exists2 || storedToken2 == nil || *storedToken2 != token2 {
		t.Errorf("Token for path2 should remain unaffected after deleting token for path1")
	}
}

// TestDeleteHostLevelCleanup verifies that deleting the last token for a host removes the host entry.
func TestDeleteHostLevelCleanup(t *testing.T) {
	store := NewInMemoryStore()
	testURL, _ := url.Parse("http://host.com/path")
	testToken := Token("token123")

	_ = store.Put(testURL, &testToken)
	_ = store.Delete(testURL)

	// Attempt to retrieve a token for the same host but different path
	otherPathURL, _ := url.Parse("http://host.com/otherpath")
	_, exists := store.Get(otherPathURL)

	if exists {
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
	testToken := Token("token123")
	_ = store.Put(baseURL, &testToken)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Delete(&url.URL{Host: "host.com", Path: "/path"})
		}()
	}

	wg.Wait()

	_, exists := store.Get(&url.URL{Host: "host.com", Path: "/path"})
	if exists {
		t.Errorf("Token should have been deleted after concurrent deletion attempts")
	}
}
