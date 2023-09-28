package registry

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

type Entry struct {
	Epoch int64  `json:"epoch"`
	Rule  string `json:"rule"`
}

var reg = map[string]Entry{}
var mu sync.Mutex

func New(s string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	new_epoch := currentEpoch()
	id := strconv.FormatInt(new_epoch, 10) + "+" + randomString(8)

	fn := "registry/" + id
	err := writeFile(fn, s)
	if err != nil {
		return "", err
	}

	reg[id] = Entry{
		Epoch: new_epoch,
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
		path = "registry/" + fn
		data, err := readFile(path)
		if err != nil {
			log.Panicf("failed to read file: %v", err)
		}

		e, _, found := strings.Cut(fn, "+")
		if !found {
			log.Panicf("invalid filename: %s", fn)
		}

		ep, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			log.Panicf("failed to parse epoch: %s", e)
		}

		reg[fn] = Entry{
			Epoch: ep,
			Rule:  data,
		}
	}

	// apply rules and dont fail if they alreay exists

	return nil
}
