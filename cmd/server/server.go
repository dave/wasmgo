package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dave/wasmgo/cmd/cmdconfig"
	"github.com/dave/wasmgo/cmd/deployer"
	"github.com/pkg/browser"
)

func Start(cfg *cmdconfig.Config) error {

	var debug io.Writer
	if cfg.Verbose {
		debug = os.Stdout
	} else {
		debug = ioutil.Discard
	}

	dep, err := deployer.New(cfg)
	if err != nil {
		return err
	}

	svr := &server{cfg: cfg, dep: dep, debug: debug}

	s := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: svr}

	go func() {
		fmt.Fprintf(debug, "Starting server on %s\n", s.Addr)
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		if cfg.Open {
			browser.OpenURL(fmt.Sprintf("http://localhost:%d/", cfg.Port))
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop

	fmt.Fprintln(debug, "Stopping server")

	return nil
}

type server struct {
	cfg   *cmdconfig.Config
	dep   *deployer.State
	debug io.Writer
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case strings.HasSuffix(req.RequestURI, "/favicon.ico"):
		// ignore
	case strings.HasSuffix(req.RequestURI, "/binary.wasm"):
		// binary
		contents, hash, err := s.dep.Build()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		w.Header().Set("Content-Type", "application/wasm")
		if _, err := io.Copy(w, bytes.NewReader(contents)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		fmt.Fprintf(s.debug, "Compiled WASM binary with hash %x\n", hash)
	case strings.HasSuffix(req.RequestURI, "/loader.js"):
		// loader js
		contents, _, err := s.dep.Loader("/binary.wasm")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := io.Copy(w, bytes.NewReader(contents)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	case strings.HasSuffix(req.RequestURI, "/script.js"):
		// script
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := io.Copy(w, bytes.NewBufferString(WasmExec)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	default:
		// index page
		contents, _, err := s.dep.Index("/script.js", "/loader.js", "/binary.wasm")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		w.Header().Set("Content-Type", "text/html")
		if _, err := io.Copy(w, bytes.NewReader(contents)); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

const WasmExec = `(()=>{if("undefined"!=typeof global);else if("undefined"!=typeof window)window.global=window;else{if("undefined"==typeof self)throw new Error("cannot export Go (neither global, window nor self is defined)");self.global=self}const e=global.process&&"node"===global.process.title;if(e){global.require=require,global.fs=require("fs");const e=require("crypto");global.crypto={getRandomValues(t){e.randomFillSync(t)}},global.performance={now(){const[e,t]=process.hrtime();return 1e3*e+t/1e6}};const t=require("util");global.TextEncoder=t.TextEncoder,global.TextDecoder=t.TextDecoder}else{let e="";global.fs={constants:{O_WRONLY:-1,O_RDWR:-1,O_CREAT:-1,O_TRUNC:-1,O_APPEND:-1,O_EXCL:-1},writeSync(t,n){const i=(e+=s.decode(n)).lastIndexOf("\n");return-1!=i&&(console.log(e.substr(0,i)),e=e.substr(i+1)),n.length},write(e,t,s,n,i,r){if(0!==s||n!==t.length||null!==i)throw new Error("not implemented");r(null,this.writeSync(e,t))},open(e,t,s,n){const i=new Error("not implemented");i.code="ENOSYS",n(i)},read(e,t,s,n,i,r){const o=new Error("not implemented");o.code="ENOSYS",r(o)},fsync(e,t){t(null)}}}const t=new TextEncoder("utf-8"),s=new TextDecoder("utf-8");if(global.Go=class{constructor(){this.argv=["js"],this.env={},this.exit=(e=>{0!==e&&console.warn("exit code:",e)}),this._exitPromise=new Promise(e=>{this._resolveExitPromise=e}),this._pendingEvent=null,this._scheduledTimeouts=new Map,this._nextCallbackTimeoutID=1;const e=()=>new DataView(this._inst.exports.mem.buffer),n=(t,s)=>{e().setUint32(t+0,s,!0),e().setUint32(t+4,Math.floor(s/4294967296),!0)},i=t=>{return e().getUint32(t+0,!0)+4294967296*e().getInt32(t+4,!0)},r=t=>{const s=e().getFloat64(t,!0);if(0===s)return;if(!isNaN(s))return s;const n=e().getUint32(t,!0);return this._values[n]},o=(t,s)=>{if("number"==typeof s)return isNaN(s)?(e().setUint32(t+4,2146959360,!0),void e().setUint32(t,0,!0)):0===s?(e().setUint32(t+4,2146959360,!0),void e().setUint32(t,1,!0)):void e().setFloat64(t,s,!0);switch(s){case void 0:return void e().setFloat64(t,0,!0);case null:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,2,!0);case!0:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,3,!0);case!1:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,4,!0)}let n=this._refs.get(s);void 0===n&&(n=this._values.length,this._values.push(s),this._refs.set(s,n));let i=0;switch(typeof s){case"string":i=1;break;case"symbol":i=2;break;case"function":i=3}e().setUint32(t+4,2146959360|i,!0),e().setUint32(t,n,!0)},l=e=>{const t=i(e+0),s=i(e+8);return new Uint8Array(this._inst.exports.mem.buffer,t,s)},a=e=>{const t=i(e+0),s=i(e+8),n=new Array(s);for(let e=0;e<s;e++)n[e]=r(t+8*e);return n},c=e=>{const t=i(e+0),n=i(e+8);return s.decode(new DataView(this._inst.exports.mem.buffer,t,n))},u=Date.now()-performance.now();this.importObject={go:{"runtime.wasmExit":t=>{const s=e().getInt32(t+8,!0);this.exited=!0,delete this._inst,delete this._values,delete this._refs,this.exit(s)},"runtime.wasmWrite":t=>{const s=i(t+8),n=i(t+16),r=e().getInt32(t+24,!0);fs.writeSync(s,new Uint8Array(this._inst.exports.mem.buffer,n,r))},"runtime.nanotime":e=>{n(e+8,1e6*(u+performance.now()))},"runtime.walltime":t=>{const s=(new Date).getTime();n(t+8,s/1e3),e().setInt32(t+16,s%1e3*1e6,!0)},"runtime.scheduleTimeoutEvent":t=>{const s=this._nextCallbackTimeoutID;this._nextCallbackTimeoutID++,this._scheduledTimeouts.set(s,setTimeout(()=>{this._resume()},i(t+8)+1)),e().setInt32(t+16,s,!0)},"runtime.clearTimeoutEvent":t=>{const s=e().getInt32(t+8,!0);clearTimeout(this._scheduledTimeouts.get(s)),this._scheduledTimeouts.delete(s)},"runtime.getRandomData":e=>{crypto.getRandomValues(l(e+8))},"syscall/js.stringVal":e=>{o(e+24,c(e+8))},"syscall/js.valueGet":e=>{const t=Reflect.get(r(e+8),c(e+16));e=this._inst.exports.getsp(),o(e+32,t)},"syscall/js.valueSet":e=>{Reflect.set(r(e+8),c(e+16),r(e+32))},"syscall/js.valueIndex":e=>{o(e+24,Reflect.get(r(e+8),i(e+16)))},"syscall/js.valueSetIndex":e=>{Reflect.set(r(e+8),i(e+16),r(e+24))},"syscall/js.valueCall":t=>{try{const s=r(t+8),n=Reflect.get(s,c(t+16)),i=a(t+32),l=Reflect.apply(n,s,i);t=this._inst.exports.getsp(),o(t+56,l),e().setUint8(t+64,1)}catch(s){o(t+56,s),e().setUint8(t+64,0)}},"syscall/js.valueInvoke":t=>{try{const s=r(t+8),n=a(t+16),i=Reflect.apply(s,void 0,n);t=this._inst.exports.getsp(),o(t+40,i),e().setUint8(t+48,1)}catch(s){o(t+40,s),e().setUint8(t+48,0)}},"syscall/js.valueNew":t=>{try{const s=r(t+8),n=a(t+16),i=Reflect.construct(s,n);t=this._inst.exports.getsp(),o(t+40,i),e().setUint8(t+48,1)}catch(s){o(t+40,s),e().setUint8(t+48,0)}},"syscall/js.valueLength":e=>{n(e+16,parseInt(r(e+8).length))},"syscall/js.valuePrepareString":e=>{const s=t.encode(String(r(e+8)));o(e+16,s),n(e+24,s.length)},"syscall/js.valueLoadString":e=>{const t=r(e+8);l(e+16).set(t)},"syscall/js.valueInstanceOf":t=>{e().setUint8(t+24,r(t+8)instanceof r(t+16))},debug:e=>{console.log(e)}}}}async run(e){this._inst=e,this._values=[NaN,0,null,!0,!1,global,this._inst.exports.mem,this],this._refs=new Map,this.exited=!1;const s=new DataView(this._inst.exports.mem.buffer);let n=4096;const i=e=>{let i=n;return new Uint8Array(s.buffer,n,e.length+1).set(t.encode(e+"\0")),n+=e.length+(8-e.length%8),i},r=this.argv.length,o=[];this.argv.forEach(e=>{o.push(i(e))});const l=Object.keys(this.env).sort();o.push(l.length),l.forEach(e=>{o.push(i(` + "`" + `${e}=${this.env[e]}` + "`" + `))});const a=n;o.forEach(e=>{s.setUint32(n,e,!0),s.setUint32(n+4,0,!0),n+=8}),this._inst.exports.run(r,a),this.exited&&this._resolveExitPromise(),await this._exitPromise}_resume(){if(this.exited)throw new Error("Go program has already exited");this._inst.exports.resume(),this.exited&&this._resolveExitPromise()}_makeFuncWrapper(e){const t=this;return function(){const s={id:e,this:this,args:arguments};return t._pendingEvent=s,t._resume(),s.result}}},e){process.argv.length<3&&(process.stderr.write("usage: go_js_wasm_exec [wasm binary] [arguments]\n"),process.exit(1));const e=new Go;e.argv=process.argv.slice(2),e.env=Object.assign({TMPDIR:require("os").tmpdir()},process.env),e.exit=process.exit,WebAssembly.instantiate(fs.readFileSync(process.argv[2]),e.importObject).then(t=>(process.on("exit",t=>{0!==t||e.exited||(e._pendingEvent={id:0},e._resume())}),e.run(t.instance))).catch(e=>{throw e})}(()=>{if("undefined"!=typeof global);else if("undefined"!=typeof window)window.global=window;else{if("undefined"==typeof self)throw new Error("cannot export Go (neither global, window nor self is defined)");self.global=self}const e=global.process&&"node"===global.process.title;if(e){global.require=require,global.fs=require("fs");const e=require("crypto");global.crypto={getRandomValues(t){e.randomFillSync(t)}},global.performance={now(){const[e,t]=process.hrtime();return 1e3*e+t/1e6}};const t=require("util");global.TextEncoder=t.TextEncoder,global.TextDecoder=t.TextDecoder}else{let e="";global.fs={constants:{O_WRONLY:-1,O_RDWR:-1,O_CREAT:-1,O_TRUNC:-1,O_APPEND:-1,O_EXCL:-1},writeSync(t,n){const i=(e+=s.decode(n)).lastIndexOf("\n");return-1!=i&&(console.log(e.substr(0,i)),e=e.substr(i+1)),n.length},write(e,t,s,n,i,r){if(0!==s||n!==t.length||null!==i)throw new Error("not implemented");r(null,this.writeSync(e,t))},open(e,t,s,n){const i=new Error("not implemented");i.code="ENOSYS",n(i)},read(e,t,s,n,i,r){const o=new Error("not implemented");o.code="ENOSYS",r(o)},fsync(e,t){t(null)}}}const t=new TextEncoder("utf-8"),s=new TextDecoder("utf-8");if(global.Go=class{constructor(){this.argv=["js"],this.env={},this.exit=(e=>{0!==e&&console.warn("exit code:",e)}),this._exitPromise=new Promise(e=>{this._resolveExitPromise=e}),this._pendingEvent=null,this._scheduledTimeouts=new Map,this._nextCallbackTimeoutID=1;const e=()=>new DataView(this._inst.exports.mem.buffer),n=(t,s)=>{e().setUint32(t+0,s,!0),e().setUint32(t+4,Math.floor(s/4294967296),!0)},i=t=>{return e().getUint32(t+0,!0)+4294967296*e().getInt32(t+4,!0)},r=t=>{const s=e().getFloat64(t,!0);if(0===s)return;if(!isNaN(s))return s;const n=e().getUint32(t,!0);return this._values[n]},o=(t,s)=>{if("number"==typeof s)return isNaN(s)?(e().setUint32(t+4,2146959360,!0),void e().setUint32(t,0,!0)):0===s?(e().setUint32(t+4,2146959360,!0),void e().setUint32(t,1,!0)):void e().setFloat64(t,s,!0);switch(s){case void 0:return void e().setFloat64(t,0,!0);case null:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,2,!0);case!0:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,3,!0);case!1:return e().setUint32(t+4,2146959360,!0),void e().setUint32(t,4,!0)}let n=this._refs.get(s);void 0===n&&(n=this._values.length,this._values.push(s),this._refs.set(s,n));let i=0;switch(typeof s){case"string":i=1;break;case"symbol":i=2;break;case"function":i=3}e().setUint32(t+4,2146959360|i,!0),e().setUint32(t,n,!0)},l=e=>{const t=i(e+0),s=i(e+8);return new Uint8Array(this._inst.exports.mem.buffer,t,s)},a=e=>{const t=i(e+0),s=i(e+8),n=new Array(s);for(let e=0;e<s;e++)n[e]=r(t+8*e);return n},c=e=>{const t=i(e+0),n=i(e+8);return s.decode(new DataView(this._inst.exports.mem.buffer,t,n))},u=Date.now()-performance.now();this.importObject={go:{"runtime.wasmExit":t=>{const s=e().getInt32(t+8,!0);this.exited=!0,delete this._inst,delete this._values,delete this._refs,this.exit(s)},"runtime.wasmWrite":t=>{const s=i(t+8),n=i(t+16),r=e().getInt32(t+24,!0);fs.writeSync(s,new Uint8Array(this._inst.exports.mem.buffer,n,r))},"runtime.nanotime":e=>{n(e+8,1e6*(u+performance.now()))},"runtime.walltime":t=>{const s=(new Date).getTime();n(t+8,s/1e3),e().setInt32(t+16,s%1e3*1e6,!0)},"runtime.scheduleTimeoutEvent":t=>{const s=this._nextCallbackTimeoutID;this._nextCallbackTimeoutID++,this._scheduledTimeouts.set(s,setTimeout(()=>{this._resume()},i(t+8)+1)),e().setInt32(t+16,s,!0)},"runtime.clearTimeoutEvent":t=>{const s=e().getInt32(t+8,!0);clearTimeout(this._scheduledTimeouts.get(s)),this._scheduledTimeouts.delete(s)},"runtime.getRandomData":e=>{crypto.getRandomValues(l(e+8))},"syscall/js.stringVal":e=>{o(e+24,c(e+8))},"syscall/js.valueGet":e=>{const t=Reflect.get(r(e+8),c(e+16));e=this._inst.exports.getsp(),o(e+32,t)},"syscall/js.valueSet":e=>{Reflect.set(r(e+8),c(e+16),r(e+32))},"syscall/js.valueIndex":e=>{o(e+24,Reflect.get(r(e+8),i(e+16)))},"syscall/js.valueSetIndex":e=>{Reflect.set(r(e+8),i(e+16),r(e+24))},"syscall/js.valueCall":t=>{try{const s=r(t+8),n=Reflect.get(s,c(t+16)),i=a(t+32),l=Reflect.apply(n,s,i);t=this._inst.exports.getsp(),o(t+56,l),e().setUint8(t+64,1)}catch(s){o(t+56,s),e().setUint8(t+64,0)}},"syscall/js.valueInvoke":t=>{try{const s=r(t+8),n=a(t+16),i=Reflect.apply(s,void 0,n);t=this._inst.exports.getsp(),o(t+40,i),e().setUint8(t+48,1)}catch(s){o(t+40,s),e().setUint8(t+48,0)}},"syscall/js.valueNew":t=>{try{const s=r(t+8),n=a(t+16),i=Reflect.construct(s,n);t=this._inst.exports.getsp(),o(t+40,i),e().setUint8(t+48,1)}catch(s){o(t+40,s),e().setUint8(t+48,0)}},"syscall/js.valueLength":e=>{n(e+16,parseInt(r(e+8).length))},"syscall/js.valuePrepareString":e=>{const s=t.encode(String(r(e+8)));o(e+16,s),n(e+24,s.length)},"syscall/js.valueLoadString":e=>{const t=r(e+8);l(e+16).set(t)},"syscall/js.valueInstanceOf":t=>{e().setUint8(t+24,r(t+8)instanceof r(t+16))},debug:e=>{console.log(e)}}}}async run(e){this._inst=e,this._values=[NaN,0,null,!0,!1,global,this._inst.exports.mem,this],this._refs=new Map,this.exited=!1;const s=new DataView(this._inst.exports.mem.buffer);let n=4096;const i=e=>{let i=n;return new Uint8Array(s.buffer,n,e.length+1).set(t.encode(e+"\0")),n+=e.length+(8-e.length%8),i},r=this.argv.length,o=[];this.argv.forEach(e=>{o.push(i(e))});const l=Object.keys(this.env).sort();o.push(l.length),l.forEach(e=>{o.push(i(` + "`" + `${e}=${this.env[e]}` + "`" + `))});const a=n;o.forEach(e=>{s.setUint32(n,e,!0),s.setUint32(n+4,0,!0),n+=8}),this._inst.exports.run(r,a),this.exited&&this._resolveExitPromise(),await this._exitPromise}_resume(){if(this.exited)throw new Error("Go program has already exited");this._inst.exports.resume(),this.exited&&this._resolveExitPromise()}_makeFuncWrapper(e){const t=this;return function(){const s={id:e,this:this,args:arguments};return t._pendingEvent=s,t._resume(),s.result}}},e){process.argv.length<3&&(process.stderr.write("usage: go_js_wasm_exec [wasm binary] [arguments]\n"),process.exit(1));const e=new Go;e.argv=process.argv.slice(2),e.env=Object.assign({TMPDIR:require("os").tmpdir()},process.env),e.exit=process.exit,WebAssembly.instantiate(fs.readFileSync(process.argv[2]),e.importObject).then(t=>(process.on("exit",t=>{0!==t||e.exited||(e._pendingEvent={id:0},e._resume())}),e.run(t.instance))).catch(e=>{throw e})}})()})();`
