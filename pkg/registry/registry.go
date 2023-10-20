package registry

import (
	"fmt"
	"log"
	"sync"
)

type Entry struct {
	Epoch int64  `json:"epoch"`
	Rule  string `json:"rule"`
}

var reg = map[string]Entry{}
var epoch Epoch
var mu sync.Mutex

func GenerateName() (string, int64) {
	e := epoch.Now()
	s := fmt.Sprintf("%d+%s", e, randomString(8))
	return s, e
}

func ParseName(id string) int64 {
	var e int64
	var s string
	_, err := fmt.Sscanf(id, "%d+%s", &e, &s)
	if err != nil {
		return 0
	}
	return e
}

func Append(id, s string) error {
	mu.Lock()
	defer mu.Unlock()

	fn := "registry/" + id
	err := writeFile(fn, s)
	if err != nil {
		return err
	}

	e := ParseName(id)
	reg[id] = Entry{
		Epoch: e,
		Rule:  s,
	}

	epoch.Update(e)
	return nil
}

func Get(id string) *Entry {
	mu.Lock()
	defer mu.Unlock()

	e, ok := reg[id]
	if ok {
		return &e
	}
	return nil
}

func Delete(id string) error {
	mu.Lock()
	defer mu.Unlock()

	e, ok := reg[id]
	if ok {
		fn := "registry/" + id
		err := deleteFile(fn)
		if err != nil {
			return err
		}

		delete(reg, id)
		epoch.Update(e.Epoch)
	}

	return nil
}

func List() []string {
	mu.Lock()
	defer mu.Unlock()

	rs := []string{}
	for id := range reg {
		rs = append(rs, id)
	}

	return rs
}

func Init(path string) error {
	mu.Lock()
	defer mu.Unlock()

	basepath = path
	epoch.SetFile(basepath + "/epoch")
	epoch.Load()

	err := makeDir("registry")
	if err != nil {
		return fmt.Errorf("registry: failed to create directory: %v", err)
	}

	files, err := listDir("registry")
	if err != nil {
		return fmt.Errorf("registry: failed to list directory: %v", err)
	}

	if epoch.Latest == 0 && len(files) > 0 {
		return fmt.Errorf("registry: found existing rules, but epoch is missing")
	}

	for _, fn := range files {
		e := ParseName(fn)
		if e == 0 {
			return fmt.Errorf("registry: invalid filename: %s", fn)
		}

		path = "registry/" + fn
		data, err := readFile(path)
		if err != nil {
			return fmt.Errorf("registry: failed to read file: %v", err)
		}

		reg[fn] = Entry{
			Epoch: e,
			Rule:  data,
		}

		log.Printf("registry: loaded file: %s", fn)
	}

	return nil
}
