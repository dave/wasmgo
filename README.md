<a href="https://patreon.com/davebrophy" title="Help with my hosting bills using Patreon"><img src="https://img.shields.io/badge/patreon-donate-yellow.svg" style="max-width:100%;"></a>

# wasmgo

The `wasmgo` command compiles Go to WASM and deploys to the [jsgo.io](https://github.com/dave/jsgo) 
CDN.

### Install

`go get -u github.com/dave/wasmgo`


### Usage

```
Compiles Go to WASM and deploys to the jsgo.io CDN.

Usage:
  wasmgo deploy [package] [flags]

Flags:
  -b, --build string      Build tags to pass to the go build command.
  -c, --command string    Name of the go command. (default "go")
  -f, --flags string      Flags to pass to the go build command.
  -h, --help              help for deploy
  -i, --index string      Specify the index page. (default "index.wasmgo.html")
  -j, --json              Return all template variables as a json blob from the deploy command.
  -o, --open              Open the page in a browser.
  -t, --template string   Template defining the output returned by the deploy command. Variables: Page (string), Loader (string). (default "{{ .Page }}")
  -v, --verbose           Show detailed status messages.
```

### Example

Here's a simple hello world:

```
wasmgo deploy github.com/dave/wasmgo/helloworld
```

### Index

You may specify a custom index page by including `index.wasmgo.html` in your project or by using the `index` 
command line flag.

Your index page should look something like this:

```html
<html>
<head><meta charset="utf-8"></head>
<body>
	<script src="{{ .Script }}"></script>
	<script src="{{ .Loader }}"></script>
</body>
</html>
```
