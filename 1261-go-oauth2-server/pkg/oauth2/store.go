package oauth2

import (
	"sync"
	"time"
)

const (
	CodeExpiry      = 5 * time.Minute
	AccessExpiry    = 1 * time.Hour
	RefreshExpiry   = 7 * 24 * time.Hour
)

type Client struct {
	ID           string
	Secret       string
	RedirectURIs []string
}

type AuthorizationCode struct {
	Code              string
	ClientID          string
	RedirectURI       string
	Scope             string
	CodeChallenge     string
	CodeChallengeMethod string
	ExpiresAt         time.Time
	Used              bool
}

type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Scope        string
	AccessExpiry time.Time
	RefreshExpiry time.Time
	ClientID     string
}

type Store struct {
	clients          sync.Map
	authCodes        sync.Map
	tokens           sync.Map
	refreshToAccess  sync.Map
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) RegisterClient(c *Client) {
	s.clients.Store(c.ID, c)
}

func (s *Store) GetClient(id string) (*Client, bool) {
	v, ok := s.clients.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Client), true
}

func (s *Store) StoreAuthCode(code *AuthorizationCode) {
	s.authCodes.Store(code.Code, code)
}

func (s *Store) GetAuthCode(code string) (*AuthorizationCode, bool) {
	v, ok := s.authCodes.Load(code)
	if !ok {
		return nil, false
	}
	return v.(*AuthorizationCode), true
}

func (s *Store) DeleteAuthCode(code string) {
	s.authCodes.Delete(code)
}

func (s *Store) StoreToken(t *Token) {
	s.tokens.Store(t.AccessToken, t)
	if t.RefreshToken != "" {
		s.refreshToAccess.Store(t.RefreshToken, t.AccessToken)
	}
}

func (s *Store) GetTokenByAccess(access string) (*Token, bool) {
	v, ok := s.tokens.Load(access)
	if !ok {
		return nil, false
	}
	return v.(*Token), true
}

func (s *Store) GetTokenByRefresh(refresh string) (*Token, bool) {
	v, ok := s.refreshToAccess.Load(refresh)
	if !ok {
		return nil, false
	}
	access := v.(string)
	return s.GetTokenByAccess(access)
}

func (s *Store) DeleteToken(access string) {
	if t, ok := s.GetTokenByAccess(access); ok {
		if t.RefreshToken != "" {
			s.refreshToAccess.Delete(t.RefreshToken)
		}
		s.tokens.Delete(access)
	}
}
