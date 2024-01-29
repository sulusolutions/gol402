package tokenstore

import (
	"net/url"
	"strings"
)

// trie is a structure for storing and searching paths in a Trie.
type trie struct {
	root *trieNode
}

// trieNode represents a node in the Trie.
type trieNode struct {
	children map[string]*trieNode
	token    *Token // Token associated with the path (only set at leaf nodes)
}

// newTrie creates a new pathTrie.
func newTrie() *trie {
	return &trie{
		root: &trieNode{
			children: make(map[string]*trieNode),
		},
	}
}

// put adds a path and its associated token to the Trie.
func (t *trie) put(rPath string, token *Token) error {
	current := t.root

	// Parse the path to extract the query as a single unit
	u, err := url.Parse(rPath)
	if err != nil {
		return err
	}

	// Extract the path and query from the URL object
	path := u.Path
	query := u.RawQuery

	// Split the path into segments
	segments := strings.Split(path, "/")

	for _, segment := range segments {
		if segment == "" {
			continue // Skip empty segments
		}

		if current.children[segment] == nil {
			current.children[segment] = &trieNode{
				children: make(map[string]*trieNode),
			}
		}
		current = current.children[segment]
	}

	// Insert the query as a single unit
	if query != "" {
		if current.children[query] == nil {
			current.children[query] = &trieNode{
				children: make(map[string]*trieNode),
			}
		}
		current = current.children[query]
	}

	current.token = token
	return nil
}

// get looks for a token that matches the given path.
// It returns the most relevant token if available; otherwise, it returns nil.
func (t *trie) get(path string) (*Token, bool) {
	current := t.root
	var closestToken *Token

	// Parse the path to extract the query as a single unit
	u, err := url.Parse(path)
	if err != nil {
		return nil, false
	}

	// Extract the path and query from the URL object
	path = u.Path
	query := u.RawQuery

	// Split the path into segments
	segments := strings.Split(path, "/")

	for _, segment := range segments {
		if segment == "" {
			continue // Skip empty segments
		}

		// Attempt to match the segment
		node := current.children[segment]
		if node == nil {
			break
		}

		current = node
		if current.token != nil {
			closestToken = current.token
		}
	}

	// If a query is present, check for a query-based token
	if query != "" {
		queryToken := current.children[query]
		if queryToken != nil && queryToken.token != nil {
			closestToken = queryToken.token
		}
	}

	return closestToken, closestToken != nil
}

// delete removes a path and its associated token from the Trie.
func (t *trie) delete(path string) error {
	current := t.root

	// Parse the path to extract the query as a single unit
	u, err := url.Parse(path)
	if err != nil {
		return err
	}

	// Extract the path and query from the URL object
	path = u.Path
	query := u.RawQuery

	// Split the path into segments
	segments := strings.Split(path, "/")

	for _, segment := range segments {
		if segment == "" {
			continue // Skip empty segments
		}

		// Attempt to match the segment
		node := current.children[segment]
		if node == nil {
			return nil
		}

		current = node
	}

	// If a query is present, check for a query-based token
	if query != "" {
		queryToken := current.children[query]
		if queryToken != nil {
			queryToken.token = nil
		}
	}

	current.token = nil
	return nil
}
