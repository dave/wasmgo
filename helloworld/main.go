package main

import "syscall/js"

func main() {
	js.Global().Get("document").Call("getElementsByTagName", "body").Index(0).Set("innerHTML", "Hello, World!")
}
