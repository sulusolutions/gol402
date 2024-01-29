package tokenstore

import (
	"testing"
)

func TestTrie_PutAndGet(t *testing.T) {
	trie := newTrie()
	token1 := Token("token1")
	token2 := Token("token2")

	// Insert paths with tokens
	paths := []struct {
		path  string
		token *Token
	}{
		{"/path1", &token1},
		{"/path2", &token2},
		{"/path/with/query?param=value", &token1},
	}

	for _, p := range paths {
		trie.put(p.path, p.token)
	}

	// Retrieve tokens
	tests := []struct {
		path   string
		token  *Token
		exists bool
	}{
		{"/path1", &token1, true},
		{"/path2", &token2, true},
		{"/path/with/query?param=value", &token1, true},
		{"/nonexistent/path", nil, false},
	}

	for _, test := range tests {
		tok, ok := trie.get(test.path)
		if ok != test.exists || tok != test.token {
			t.Errorf("Expected token %v for path %s, got %v", test.token, test.path, tok)
		}
	}
}

func TestTrie_Delete(t *testing.T) {
	// Create a new pathTrie
	trie := newTrie()

	// Insert tokens
	token1 := Token("token1")
	token2 := Token("token2")
	trie.put("/path1/path2", &token1)
	trie.put("/path1/path2?query1", &token2)

	tests := []struct {
		path      string
		wantErr   bool
		wantToken bool
	}{
		{"/path1/path2", false, false},       // Delete existing path
		{"path1/path2?query1", false, false}, // Delete existing path with query
		{"/non-existing-path", false, false}, // Delete non-existing path
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			err := trie.delete(test.path)

			if gotError := err != nil; gotError != test.wantErr {
				t.Errorf("Error deleting path: wantErr = %v, gotErr = %v", test.wantErr, gotError)
			}

			_, gotToken := trie.get(test.path)
			if gotToken != test.wantToken {
				t.Errorf("Error getting token after deletion: wantToken = %v, gotToken = %v", test.wantToken, gotToken)
			}
		})
	}
}
