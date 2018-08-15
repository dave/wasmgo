// +build dev,!local

package config

const (
	LOCAL = false
)

var Host = map[string]string{
	Wasm:  "localhost:8083",
	Pkg:   "dev-pkg.jsgo.io",
	Index: "dev-index.jsgo.io",
}

var Protocol = map[string]string{
	Wasm:  "http",
	Pkg:   "https",
	Index: "https",
}
