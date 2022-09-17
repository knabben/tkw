package config

import (
	"bufio"
	"os"
	"strings"
	"tkw/pkg/template"
)

// ConvertConfigIntoMap returns a mapping for a key value configuration file
func ConvertConfigIntoMap(path string) (*template.Mapper, error) {
	var mappers template.Mapper = map[string]string{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint

	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		key, value, found := strings.Cut(fileScanner.Text(), ":")
		if found {
			mappers.Set(key, strings.TrimSpace(value))
		}
	}
	return &mappers, nil
}
