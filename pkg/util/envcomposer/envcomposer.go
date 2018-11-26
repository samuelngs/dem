package envcomposer

import "fmt"

type envmap map[string]string

// Composer for environment variables
type Composer interface {
	Set(string, string)
	Del(string)
	AsArray() []string
	AsMap() map[string]string
}

func (v envmap) Set(key, val string) {
	v[key] = val
}

func (v envmap) Del(key string) {
	delete(v, key)
}

func (v envmap) AsArray() []string {
	res := make([]string, len(v))
	var i int
	for key, val := range v {
		res[i] = fmt.Sprintf("%s=%s", key, val)
		i++
	}
	return res
}

func (v envmap) AsMap() map[string]string {
	return v
}

// New initializes an instance of environment composer
func New() Composer {
	composer := make(envmap)
	return composer
}
