package config

import (
	"bufio"
	"os"
	"strings"
)

type Mapper map[string]string

func (m Mapper) Get(key string) string {
	return m[key]
}

func (m Mapper) Set(key, value string) {
	m[key] = value
}

// LoadConfig returns the mapper object from config
func LoadConfig(config string) (mapper *Mapper, err error) {
	if mapper, err = ConvertConfigIntoMap(config); err != nil {
		return nil, err
	}
	return
}

// ConvertConfigIntoMap returns a mapping for a key value configuration file
func ConvertConfigIntoMap(path string) (*Mapper, error) {
	var mappers Mapper = map[string]string{}

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
