package tokenstore

import "net/url"

type NoOpStore struct{}

func NewNoopStore() Store {
	return &NoOpStore{}
}

func (s *NoOpStore) Put(u *url.URL, token Token) error {
	return nil
}

func (s *NoOpStore) Get(u *url.URL) (Token, bool) {
	return "", false
}

func (s *NoOpStore) Delete(u *url.URL) error {
	return nil
}
