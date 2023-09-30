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
var mu sync.Mutex

func GenerateName() (string, int64) {
	e := currentEpoch()
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

func Append(id, s string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	fn := "registry/" + id
	err := writeFile(fn, s)
	if err != nil {
		return "", err
	}

	e := ParseName(id)
	reg[id] = Entry{
		Epoch: e,
		Rule:  s,
	}

	updateEpoch()
	return id, nil
}

func Get(id string) string {
	mu.Lock()
	defer mu.Unlock()

	e, ok := reg[id]
	if ok {
		return e.Rule
	}
	return ""
}

func Delete(id string) error {
	mu.Lock()
	defer mu.Unlock()

	_, ok := reg[id]
	if ok {
		fn := "registry/" + id
		err := deleteFile(fn)
		if err != nil {
			return err
		}
		delete(reg, id)
	}

	updateEpoch()
	return nil
}

func Init(path string) error {
	mu.Lock()
	defer mu.Unlock()

	basepath = path
	readEpoch()

	err := makeDir("registry")
	if err != nil {
		log.Panicf("failed to create directory: %v", err)
	}

	files, err := listDir("registry")
	if err != nil {
		log.Panicf("failed to list directory: %v", err)
	}

	if epoch == 0 && len(files) > 0 {
		log.Panic("found existing rules, but epoch was not set")
	}

	for _, fn := range files {
		e := ParseName(fn)
		if e == 0 {
			log.Panicf("invalid filename: %s", fn)
		}

		path = "registry/" + fn
		data, err := readFile(path)
		if err != nil {
			log.Panicf("failed to read file: %v", err)
		}

		reg[fn] = Entry{
			Epoch: e,
			Rule:  data,
		}
	}

	// apply rules and dont fail if they alreay exists

	return nil
}
