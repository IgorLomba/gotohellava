package ava

import "sync"

type Links struct {
	mu sync.RWMutex
	m  map[string]string
}

func NewLinks() *Links {
	return &Links{
		m: make(map[string]string),
	}
}

func (lm *Links) Get(key string) (string, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	val, ok := lm.m[key]
	return val, ok
}

func (lm *Links) Set(key, value string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.m[key] = value
}
