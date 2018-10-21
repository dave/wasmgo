<a href="https://patreon.com/davebrophy" title="Help with my hosting bills using Patreon"><img src="https://img.shields.io/badge/patreon-donate-yellow.svg" style="max-width:100%;"></a>

# wasmgo

The `wasmgo` command compiles Go to WASM, and serves the binary locally or deploys to the [jsgo.io](https://github.com/dave/jsgo) CDN.

### Install

```
go get -u github.com/dave/wasmgo
```

### Serve command

```
wasmgo serve [flags] [package]
```

Serves the WASM with a local web server (default to port 8080). Refresh the browser to recompile.

### Deploy command

```
wasmgo deploy [flags] [package]
```

Deploys the WASM to the [jsgo.io](https://github.com/dave/jsgo) CDN.

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

### Package

Omit the package argument to use the code in the current directory.

### Examples

Here's a simple hello world:

```
wasmgo serve github.com/dave/wasmgo/helloworld
```

The page (http://localhost:8080/) opens in a browser.

Here's an amazing 2048 clone from [hajimehoshi](https://github.com/hajimehoshi):

```
go get -u github.com/hajimehoshi/ebiten/examples/2048/...
wasmgo deploy -b=example github.com/hajimehoshi/ebiten/examples/2048
```

[The deployed page](https://jsgo.io/2893575ab26da60ef14801541b46201c9d54db13) opens in a browser.

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

The index page template and the `-t` flag are both Go templates with several variables available:

* *Page*  
  The URL of the page on jsgo.io (deploy command output only).  
  
* *Script*  
  To load and execute a WASM binary, a some JS bootstrap code is required. The wasmgo command uses a minified 
  version of the [example in the official Go repo](https://github.com/golang/go/blob/master/misc/wasm/wasm_exec.js). 
  The URL of this script is the `Script` template variable.  

* *Loader*  
  The loader JS is a simple script that loads and executes the WASM binary. It's based on the [example 
  in the official Go repo](https://github.com/golang/go/blob/master/misc/wasm/wasm_exec.html#L17-L36), 
  but simplified to execute the program immediately instead. The URL of this script is the `Loader` template variable.         
  
* *Binary*  
  The URL of the WASM binary file.  

### Static files

Unfortunately wasmgo does not host your static files. I recommend using [rawgit.com](https://rawgit.com/) 
to serve static files. 

### Package splitting

The `wasmgo deploy` can't yet split the binary output by Go package, [can you help?](https://github.com/dave/wasmgo/issues/2) 
