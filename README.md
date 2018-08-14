<a href="https://patreon.com/davebrophy" title="Help with my hosting bills using Patreon"><img src="https://img.shields.io/badge/patreon-donate-yellow.svg" style="max-width:100%;"></a>

# wasmgo

The `wasmgo` command compiles Go to WASM and deploys to the [jsgo.io](https://github.com/dave/jsgo) 
CDN.

```
Compile Go to WASM, test locally or deploy to jsgo.io

Usage:
  wasmgo [command]

Available Commands:
  deploy      Compile and deploy
  help        Help about any command
  serve       Serve locally
  version     Show client version number

Flags:
  -c, --command string          Name of the go command. (default "go")
  -f, --flags string            Flags to pass to the go build command.
  -h, --help                    help for wasmgo
  -i, --index index.jsgo.html   Specify the index page. If omitted, use index.jsgo.html if it exists. (default "index.jsgo.html")
  -o, --open                    Open the page in a browser.
  -v, --verbose                 Show detailed status messages.
```

### Deploy

```
Compiles Go to WASM and deploys to the jsgo.io CDN.

Usage:
  wasmgo deploy [package] [flags]

Flags:
  -h, --help              help for deploy
  -j, --json              Return all template variables as a json blob from the deploy command.
  -t, --template string   Template defining the output returned by the deploy command. Variables: Page (string), Loader (string). (default "{{ .Page }}")
```

### Index

You may specify a custom index page by including `index.jsgo.html` in your project or by using the `index` 
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


### Serve

Serve mode coming soon.