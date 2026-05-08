package regexengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	DefaultTimeout = 5 * time.Second
)

var (
	ErrTimeout       = errors.New("operation timeout")
	ErrRegexNotFound = errors.New("regex not found")
)

type Engine struct {
	cache        sync.Map
	singleflight singleflight.Group
}

type matchResult struct {
	matches [][]int
	err     error
}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Compile(ctx context.Context, pattern string) (string, error) {
	if len(pattern) == 0 {
		return "", errors.New("pattern is empty")
	}

	regexID := generateRegexID(pattern)

	if v, ok := e.cache.Load(regexID); ok {
		if _, valid := v.(*regexp.Regexp); valid {
			return regexID, nil
		}
	}

	result, err, _ := e.singleflight.Do(regexID, func() (interface{}, error) {
		if v, ok := e.cache.Load(regexID); ok {
			if _, valid := v.(*regexp.Regexp); valid {
				return v, nil
			}
		}

		compileResult := make(chan *regexp.Regexp, 1)
		compileErr := make(chan error, 1)

		go func() {
			defer close(compileResult)
			defer close(compileErr)
			re, err := regexp.Compile(pattern)
			if err != nil {
				compileErr <- err
				return
			}
			compileResult <- re
		}()

		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case err := <-compileErr:
			return nil, err
		case re := <-compileResult:
			e.cache.Store(regexID, re)
			return re, nil
		}
	})

	if err != nil {
		return "", err
	}

	_, ok := result.(*regexp.Regexp)
	if !ok {
		return "", errors.New("compilation failed")
	}

	return regexID, nil
}

func (e *Engine) Match(ctx context.Context, regexID string, text string) ([][]int, error) {
	if len(regexID) == 0 {
		return nil, errors.New("regex_id is empty")
	}

	v, ok := e.cache.Load(regexID)
	if !ok {
		return nil, ErrRegexNotFound
	}

	re, ok := v.(*regexp.Regexp)
	if !ok {
		return nil, ErrRegexNotFound
	}

	matchResultCh := make(chan matchResult, 1)

	go func() {
		defer close(matchResultCh)
		matches := re.FindAllSubmatchIndex([]byte(text), -1)
		matchResultCh <- matchResult{matches: matches, err: nil}
	}()

	select {
	case <-ctx.Done():
		return nil, ErrTimeout
	case result := <-matchResultCh:
		return result.matches, result.err
	}
}

func (e *Engine) Purge() int {
	count := 0
	e.cache.Range(func(key, value interface{}) bool {
		e.cache.Delete(key)
		count++
		return true
	})
	return count
}

func generateRegexID(pattern string) string {
	hash := sha256.Sum256([]byte(pattern))
	return hex.EncodeToString(hash[:])
}
