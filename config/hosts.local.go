// +build local

package config

const (
	LOCAL = true
)

var Host = map[string]string{
	Wasm:  "localhost:8083",
	Pkg:   "localhost:8092",
	Index: "localhost:8093",
}

var Protocol = map[string]string{
	Wasm:  "http",
	Pkg:   "http",
	Index: "http",
}
