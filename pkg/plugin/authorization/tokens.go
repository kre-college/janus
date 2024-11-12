package authorization

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hellofresh/janus/pkg/config"
)

type JWTToken struct {
	ID             uint64    `json:"id"`
	UserID         uint64    `json:"user_id"`
	Token          string    `json:"token"`
	ExpirationDate time.Time `json:"expiration_date"`
}

type TokenManager struct {
	Tokens map[string]*Claims
	Conf   *config.Authorization
	mu     sync.RWMutex
}

func NewTokenManager(conf *config.Authorization) *TokenManager {
	return &TokenManager{
		Tokens: make(map[string]*Claims),
		Conf:   conf,
	}
}

func (tm *TokenManager) FetchTokensWithRetry(retryAttempts int, retryTimeout time.Duration) error {
	for i := 0; i < retryAttempts; i++ {
		err := tm.FetchTokens()
		if err != nil {
			t := time.NewTimer(retryTimeout)
			<-t.C
			continue
		}
		return nil
	}

	return fmt.Errorf("unable to fetch tokens after %d attempts", retryAttempts)
}

func (tm *TokenManager) FetchTokens() error {
	url := fmt.Sprintf("%s/%s/tokens", tm.Conf.UserManagementURL, tm.Conf.ApiVersion)
	body, err := doGetRequestWithTimeout(url, 3*time.Second)
	if err != nil {
		if errors.Is(err, ErrTimeout) {
			return nil
		}
		return err
	}

	tokensSlice := make([]*JWTToken, 0)
	err = json.Unmarshal(body, &tokensSlice)
	if err != nil {
		return err
	}

	tokensMap, err := tokenSliceToStringClaimsMap(tokensSlice)
	if err != nil {
		return err
	}

	tm.Tokens = tokensMap

	return nil
}

func tokenSliceToStringClaimsMap(tokensSlice []*JWTToken) (map[string]*Claims, error) {
	tokensMap := make(map[string]*Claims)

	for _, token := range tokensSlice {
		claims, err := ExtractClaims(token.Token)
		if err != nil {
			return nil, err
		}
		tokensMap[token.Token] = claims
	}

	return tokensMap, nil
}

func (tm *TokenManager) UpsertToken(token string) error {
	claims, err := ExtractClaims(token)
	if err != nil {
		return err
	}

	tm.Tokens[token] = claims

	go tm.deleteTokenAfterExpiration(token)

	return nil
}

func (tm *TokenManager) DeleteToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.Tokens, token)
}

func (tm *TokenManager) isTokenValid(accessToken string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if _, exists := tm.Tokens[accessToken]; exists {
		return true
	}

	return false
}

func (tm *TokenManager) deleteTokenAfterExpiration(token string) {
	claims, exists := tm.Tokens[token]
	if !exists {
		return
	}

	expiresAt := time.Unix(claims.ExpiresAt, 0)
	duration := time.Until(expiresAt)

	timer := time.NewTimer(duration)
	<-timer.C

	tm.DeleteToken(token)
}
