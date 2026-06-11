package store

import (
	"strings"
	"sync"
	"time"
)

type DataPoint struct {
	Value     string    `json:"value"`
	Numeric   *float64  `json:"numeric,omitempty"`
	Unit      string    `json:"unit,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RawName   string    `json:"raw_name"`
}

type Store struct {
	mu   sync.RWMutex
	data map[string]DataPoint
}

func New() *Store {
	return &Store{data: make(map[string]DataPoint)}
}

func (s *Store) Set(name string, dp DataPoint) {
	s.mu.Lock()
	s.data[name] = dp
	s.mu.Unlock()
}

func (s *Store) Get(name string) (DataPoint, bool) {
	s.mu.RLock()
	dp, ok := s.data[name]
	s.mu.RUnlock()
	return dp, ok
}

func (s *Store) GetAll() map[string]DataPoint {
	s.mu.RLock()
	result := make(map[string]DataPoint, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	s.mu.RUnlock()
	return result
}

func (s *Store) GetByPrefix(prefix string) map[string]DataPoint {
	s.mu.RLock()
	result := make(map[string]DataPoint)
	for k, v := range s.data {
		if strings.HasPrefix(k, prefix) {
			result[k] = v
		}
	}
	s.mu.RUnlock()
	return result
}

func (s *Store) Count() int {
	s.mu.RLock()
	n := len(s.data)
	s.mu.RUnlock()
	return n
}

func (s *Store) LatestTimestamp() time.Time {
	s.mu.RLock()
	var latest time.Time
	for _, dp := range s.data {
		if dp.Timestamp.After(latest) {
			latest = dp.Timestamp
		}
	}
	s.mu.RUnlock()
	return latest
}
