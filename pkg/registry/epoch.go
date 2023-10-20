package registry

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Epoch struct {
	fn       string
	mu       sync.Mutex
	Latest   int64 `json:"latest"`
	Checksum int64 `json:"checksum"`
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
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return os.WriteFile(e.fn, data, 0640)
}

func (e *Epoch) Now() int64 {
	return time.Now().UnixMilli()
}

func (e *Epoch) Update(val int64) {
	e.mu.Lock()
	e.Checksum += val
	if val > e.Latest {
		e.Latest = val
	}
	e.mu.Unlock()
	e.Store()
}

func (e *Epoch) SetFile(fn string) {
	e.fn = fn
}
