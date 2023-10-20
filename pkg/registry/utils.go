package registry

import (
	"math/rand"
	"os"
	"sort"
	"strings"
)

var basepath string

func randomString(n int) string {
	var sb strings.Builder
	var alphabet []rune = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	size := len(alphabet)

	for i := 0; i < n; i++ {
		ch := alphabet[rand.Intn(size)]
		sb.WriteRune(ch)
	}

	return sb.String()
}

func readFile(fn string) (string, error) {
	path := basepath + "/" + fn
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeFile(fn string, data string) error {
	path := basepath + "/" + fn
	err := os.WriteFile(path, []byte(data), 0640)
	if err != nil {
		return err
	}
	return nil
}

func deleteFile(fn string) error {
	path := basepath + "/" + fn
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func makeDir(fn string) error {
	path := basepath + "/" + fn
	err := os.MkdirAll(path, 0750)
	if err != nil {
		return err
	}
	return nil
}

func listDir(fn string) ([]string, error) {
	var files []string

	path := basepath + "/" + fn
	list, err := os.ReadDir(path)
	if err != nil {
		return files, err
	}

	for _, file := range list {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}
