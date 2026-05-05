package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"go-mailing-list/pkg/protocol"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	if len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

func replaceVariables(content string, subscriber *protocol.Subscriber) string {
	result := content
	result = strings.ReplaceAll(result, "{{name}}", subscriber.Name)
	result = strings.ReplaceAll(result, "{{email}}", subscriber.Email)
	result = strings.ReplaceAll(result, "{{registered_at}}", subscriber.RegisteredAt.Format(time.RFC3339))

	for key, value := range subscriber.CustomFields {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

func splitIntoBatches[T any](items []T, batchSize int) [][]T {
	if batchSize <= 0 {
		batchSize = protocol.MaxRecipientsPerBatch
	}

	var batches [][]T
	totalItems := len(items)
	numBatches := int(math.Ceil(float64(totalItems) / float64(batchSize)))

	for i := 0; i < numBatches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > totalItems {
			end = totalItems
		}
		batches = append(batches, items[start:end])
	}

	return batches
}

func shuffleSlice[T any](slice []T) {
	for i := len(slice) - 1; i > 0; i-- {
		j := int64(0)
		randBytes := make([]byte, 8)
		rand.Read(randBytes)
		for k := 0; k < 8; k++ {
			j = (j << 8) | int64(randBytes[k])
		}
		j = j % int64(i+1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *SafeCounter) Dec() {
	c.mu.Lock()
	c.value--
	c.mu.Unlock()
}

func (c *SafeCounter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func parseCSVLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	var result []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		char := line[i]
		switch char {
		case ',':
			if inQuotes {
				current.WriteByte(char)
			} else {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			}
		case '"':
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				current.WriteByte('"')
				i++
			} else {
				inQuotes = !inQuotes
			}
		default:
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}
