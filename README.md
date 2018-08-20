<a href="https://patreon.com/davebrophy" title="Help with my hosting bills using Patreon"><img src="https://img.shields.io/badge/patreon-donate-yellow.svg" style="max-width:100%;"></a>

# wasmgo

The `wasmgo` command compiles Go to WASM, and serves the binary locally or deploys to the [jsgo.io](https://github.com/dave/jsgo) CDN.

### Install

`go get -u github.com/dave/wasmgo`

### Serve command

```
  wasmgo serve [flags] [package]
```

Serves the WASM with a local web server (default to port 8080). Refresh the browser to recompile.

### Deploy command

```
  wasmgo deploy [flags] [package]
```

Deploys the WASM to jsgo.io.

### Global flags

```
  -b, --build string     Build tags to pass to the go build command.
  -c, --command string   Name of the go command. (default "go")
  -f, --flags string     Flags to pass to the go build command.
  -h, --help             help for wasmgo
  -i, --index string     Specify the index page template. Variables: Script, Loader, Binary. (default "index.wasmgo.html")
  -o, --open             Open the page in a browser. (default true)
  -v, --verbose          Show detailed status messages.
```

### Deploy flags

```
  -j, --json              Return all template variables as a json blob from the deploy command.
  -t, --template string   Template defining the output returned by the deploy command. Variables: Page, Script, Loader, Binary. (default "{{ .Page }}")
```

### Serve flags

```
  -p, --port int   Server port. (default 8080)
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

### Template variables

The index.wasmgo.html template and the `-t` flag are both templates with several variables available:

* Page - the URL of the page on jsgo.io (command output only)  
* Script - the URL of the wasm_exec.js file  
* Loader - the URL of the loader js  
* Binary - the URL of the WASM binary  
