// +build !dev,!local

package config

const (
	LOCAL = false
)

var Host = map[string]string{
	Wasm:  "wasm.jsgo.io",
	Pkg:   "pkg.jsgo.io",
	Index: "jsgo.io",
}

var Protocol = map[string]string{
	Wasm:  "https",
	Pkg:   "https",
	Index: "https",
}
