package deployer

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/jsgo/server/wasm/messages"
	"github.com/dave/services/constor/constormsg"
	"github.com/dave/wasmgo/cmd/cmdconfig"
	"github.com/dave/wasmgo/cmd/config"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
)

const CLIENT_VERSION = "1.0.0"

func New(cfg *cmdconfig.Config) (*State, error) {
	sourceDir, err := runGoList(cfg)
	if err != nil {
		return nil, err
	}
	s := &State{cfg: cfg, dir: sourceDir}
	if cfg.Verbose {
		s.debug = os.Stdout
	} else {
		s.debug = ioutil.Discard
	}
	return s, nil
}

type State struct {
	cfg   *cmdconfig.Config
	dir   string
	debug io.Writer
}

func (d *State) Start() error {

	fmt.Fprintln(d.debug, "Compiling...")

	binaryBytes, binaryHash, err := d.Build()
	if err != nil {
		return err
	}

	files := map[messages.DeployFileType]messages.DeployFile{}

	files[messages.DeployFileTypeWasm] = messages.DeployFile{
		DeployFileKey: messages.DeployFileKey{
			Type: messages.DeployFileTypeWasm,
			Hash: fmt.Sprintf("%x", binaryHash),
		},
		Contents: binaryBytes,
	}

	binaryUrl := fmt.Sprintf("%s://%s/%x.wasm", config.Protocol[config.Pkg], config.Host[config.Pkg], binaryHash)
	loaderBytes, loaderHash, err := d.Loader(binaryUrl)

	files[messages.DeployFileTypeLoader] = messages.DeployFile{
		DeployFileKey: messages.DeployFileKey{
			Type: messages.DeployFileTypeLoader,
			Hash: fmt.Sprintf("%x", loaderHash),
		},
		Contents: loaderBytes,
	}

	loaderUrl := fmt.Sprintf("%s://%s/%x.js", config.Protocol[config.Pkg], config.Host[config.Pkg], loaderHash)
	scriptUrl := fmt.Sprintf("%s://%s/wasm_exec.%s.js", config.Protocol[config.Pkg], config.Host[config.Pkg], std.Wasm[true])

	indexBytes, indexHash, err := d.Index(scriptUrl, loaderUrl, binaryUrl)
	if err != nil {
		return err
	}

	files[messages.DeployFileTypeIndex] = messages.DeployFile{
		DeployFileKey: messages.DeployFileKey{
			Type: messages.DeployFileTypeIndex,
			Hash: fmt.Sprintf("%x", indexHash),
		},
		Contents: indexBytes,
	}

	indexUrl := fmt.Sprintf("%s://%s/%x", config.Protocol[config.Index], config.Host[config.Index], indexHash)

	message := messages.DeployQuery{
		Version: CLIENT_VERSION,
		Files: []messages.DeployFileKey{
			files[messages.DeployFileTypeWasm].DeployFileKey,
			files[messages.DeployFileTypeIndex].DeployFileKey,
			files[messages.DeployFileTypeLoader].DeployFileKey,
		},
	}

	fmt.Fprintln(d.debug, "Querying server...")

	protocol := "wss"
	if config.Protocol[config.Wasm] == "http" {
		protocol = "ws"
	}
	conn, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("%s://%s/_wasm/", protocol, config.Host[config.Wasm]),
		http.Header{"Origin": []string{fmt.Sprintf("%s://%s/", config.Protocol[config.Wasm], config.Host[config.Wasm])}},
	)
	if err != nil {
		return err
	}

	messageBytes, messageType, err := messages.Marshal(message)
	if err != nil {
		return err
	}
	if err := conn.WriteMessage(messageType, messageBytes); err != nil {
		return err
	}

	var response messages.DeployQueryResponse
	var done bool
	for !done {
		_, replyBytes, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		message, err := messages.Unmarshal(replyBytes)
		if err != nil {
			return err
		}
		switch message := message.(type) {
		case messages.DeployQueryResponse:
			response = message
			done = true
		case servermsg.Queueing:
			// don't print
		case servermsg.Error:
			return errors.New(message.Message)
		case messages.DeployClientVersionNotSupported:
			return errors.New("this client version is not supported - try `go get -u github.com/dave/wasmgo`")
		default:
			// unexpected
			fmt.Fprintf(d.debug, "Unexpected message from server: %#v\n", message)
		}
	}

	if len(response.Required) > 0 {

		fmt.Fprintf(d.debug, "Files required: %d.\n", len(response.Required))
		fmt.Fprintln(d.debug, "Bundling required files...")

		var required []messages.DeployFile
		for _, k := range response.Required {
			file := files[k.Type]
			if file.Hash == "" {
				return errors.New("server requested file not found")
			}
			required = append(required, file)
		}

		payload := messages.DeployPayload{Files: required}
		payloadBytes, payloadType, err := messages.Marshal(payload)
		if err != nil {
			return err
		}
		if err := conn.WriteMessage(payloadType, payloadBytes); err != nil {
			return err
		}

		fmt.Fprintf(d.debug, "Sending payload: %dKB.\n", len(payloadBytes)/1024)

		done = false
		for !done {
			_, replyBytes, err := conn.ReadMessage()
			if err != nil {
				return err
			}
			message, err := messages.Unmarshal(replyBytes)
			if err != nil {
				return err
			}
			switch message := message.(type) {
			case messages.DeployDone:
				done = true
			case servermsg.Queueing:
				// don't print
			case servermsg.Error:
				return errors.New(message.Message)
			case constormsg.Storing:
				if message.Remain > 0 || message.Finished > 0 {
					fmt.Fprintf(d.debug, "Storing, %d to go.\n", message.Remain)
				}
			default:
				// unexpected
				fmt.Fprintf(d.debug, "Unexpected message from server: %#v\n", message)
			}
		}

		fmt.Fprintln(d.debug, "Sending done.")

	} else {
		fmt.Fprintln(d.debug, "No files required.")
	}

	outputVars := struct{ Page, Script, Loader, Binary string }{
		Page:   indexUrl,
		Script: scriptUrl,
		Loader: loaderUrl,
		Binary: binaryUrl,
	}

	if d.cfg.Json {
		out, err := json.Marshal(outputVars)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {
		tpl, err := template.New("main").Parse(d.cfg.Template)
		if err != nil {
			return err
		}
		tpl.Execute(os.Stdout, outputVars)
		fmt.Println("")
	}

	if d.cfg.Open {
		browser.OpenURL(indexUrl)
	}

	return nil
}

func (d *State) Loader(binaryUrl string) (contents, hash []byte, err error) {
	loaderBuf := &bytes.Buffer{}
	loaderSha := sha1.New()
	loaderVars := struct{ Binary string }{
		Binary: binaryUrl,
	}
	if err := loaderTemplateMin.Execute(io.MultiWriter(loaderBuf, loaderSha), loaderVars); err != nil {
		return nil, nil, err
	}
	return loaderBuf.Bytes(), loaderSha.Sum(nil), nil
}

func (d *State) Index(scriptUrl, loaderUrl, binaryUrl string) (contents, hash []byte, err error) {
	indexBuf := &bytes.Buffer{}
	indexSha := sha1.New()
	indexVars := struct{ Script, Loader, Binary string }{
		Script: scriptUrl,
		Loader: loaderUrl,
		Binary: binaryUrl,
	}
	indexTemplate := defaultIndexTemplate
	if d.cfg.Index != "" {
		indexFilename := d.cfg.Index
		if d.cfg.Path != "" {
			indexFilename = filepath.Join(d.dir, d.cfg.Index)
		}
		indexTemplateBytes, err := ioutil.ReadFile(indexFilename)
		if err != nil && !os.IsNotExist(err) {
			return nil, nil, err
		}
		if err == nil {
			indexTemplate, err = template.New("main").Parse(string(indexTemplateBytes))
			if err != nil {
				return nil, nil, err
			}
		}
	}
	if err := indexTemplate.Execute(io.MultiWriter(indexBuf, indexSha), indexVars); err != nil {
		return nil, nil, err
	}
	return indexBuf.Bytes(), indexSha.Sum(nil), nil
}

func (d *State) Build() (contents, hash []byte, err error) {

	// create a temp dir
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(tempDir)

	fpath := filepath.Join(tempDir, "out.wasm")

	args := []string{"build", "-o", fpath}

	extraFlags := strings.Fields(d.cfg.Flags)
	for _, f := range extraFlags {
		args = append(args, f)
	}

	if d.cfg.BuildTags != "" {
		args = append(args, "-tags", d.cfg.BuildTags)
	}

	path := "."
	if d.cfg.Path != "" {
		path = d.cfg.Path
	}
	args = append(args, path)

	cmd := exec.Command(d.cfg.Command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOARCH=wasm")
	cmd.Env = append(cmd.Env, "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "unsupported GOOS/GOARCH pair js/wasm") {
			return nil, nil, errors.New("you need Go v1.11 to compile WASM. It looks like your default `go` command is not v1.11. Perhaps you need the -c flag to specify a custom command name - e.g. `-c=go1.11beta3`")
		}
		return nil, nil, fmt.Errorf("%v: %s", err, string(output))
	}
	if len(output) > 0 {
		return nil, nil, fmt.Errorf("%s", string(output))
	}

	binaryBytes, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, nil, err
	}
	binarySha := sha1.New()
	if _, err := io.Copy(binarySha, bytes.NewBuffer(binaryBytes)); err != nil {
		return nil, nil, err
	}

	return binaryBytes, binarySha.Sum(nil), nil
}

func runGoList(cfg *cmdconfig.Config) (string, error) {
	args := []string{"list"}

	if cfg.BuildTags != "" {
		args = append(args, "-tags", cfg.BuildTags)
	}

	args = append(args, "-f", "{{.Dir}}")

	path := "."
	if cfg.Path != "" {
		path = cfg.Path
	}
	args = append(args, path)

	cmd := exec.Command(cfg.Command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOARCH=wasm")
	cmd.Env = append(cmd.Env, "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "unsupported GOOS/GOARCH pair js/wasm") {
			return "", errors.New("you need Go v1.11 to compile WASM. It looks like your default `go` command is not v1.11. Perhaps you need the -c flag to specify a custom command name - e.g. `-c=go1.11beta3`")
		}
		return "", fmt.Errorf("%v: %s", err, string(output))
	}
	return string(output), nil
}
