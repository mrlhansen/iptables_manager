package registry

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type Epoch struct {
	fn      string
	mu      sync.Mutex
	Current int64 `json:"current"`
}

func (e *Epoch) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := os.ReadFile(e.fn)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, e)
}

func (e *Epoch) Store() error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return os.WriteFile(e.fn, data, 0640)
}

func (e *Epoch) Now() int64 {
	return time.Now().UnixMilli()
}

func (e *Epoch) Update(ep int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ep == 0 {
		ep = e.Now()
	}
	if ep > e.Current {
		e.Current = ep
	}
	e.Store()

	log.Printf("registry: epoch updated: epoch=%d", e.Current)
}

func (e *Epoch) SetFile(fn string) {
	e.fn = fn
}
