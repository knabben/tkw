package template

type Mapper map[string]string

func (m Mapper) Get(key string) string {
	return m[key]
}

func (m Mapper) Set(key, value string) {
	m[key] = value
}
