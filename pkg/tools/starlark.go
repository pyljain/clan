package tools

import (
	"bytes"
	"clan/pkg/llm"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type starlarkHandler struct {
	definition *StarlarkTool
}

func NewStarlarkHandler(def *StarlarkTool) Tool {
	return &starlarkHandler{
		definition: def,
	}
}

func (r *starlarkHandler) Name() string {
	return r.definition.Name
}

func (r *starlarkHandler) Schema() llm.Tool {

	properties := map[string]interface{}{}
	for _, p := range r.definition.Parameters {
		properties[p.Name] = map[string]interface{}{
			"type":        p.Type,
			"description": p.Description,
		}
	}

	return llm.Tool{
		Name:        r.Name(),
		Description: r.definition.Description,
		Schema: map[string]interface{}{
			"type":       "object",
			"properties": properties,
		},
	}
}

func (r *starlarkHandler) Execute(params map[string]interface{}) (string, error) {
	sf := r.definition.Function
	thread := &starlark.Thread{Name: "function thread"}

	predeclared := starlark.StringDict{
		"get": starlark.NewBuiltin("get", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			log.Printf("Called function with args %s", args)
			req, err := http.NewRequest(http.MethodGet, args[0].(starlark.String).GoString(), nil)
			if err != nil {
				return starlark.String(err.Error()), err
			}

			client := http.Client{}
			res, err := client.Do(req)
			if err != nil {
				return starlark.String(err.Error()), err
			}

			defer res.Body.Close()

			resBytes, err := io.ReadAll(res.Body)
			if err != nil {
				return starlark.String(err.Error()), err
			}

			return starlark.String(string(resBytes)), nil
		}),

		"post": starlark.NewBuiltin("post", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			log.Printf("Called function with args %s", args)
			url := args[0].(starlark.String).GoString()
			headers := args[1].(*starlark.Dict)
			body := args[2].(starlark.String).GoString()
			log.Printf("body is %s and url is %s", body, url)

			buf := bytes.NewBufferString(body)
			req, err := http.NewRequest(http.MethodPost, url, buf)
			if err != nil {
				return starlark.String(err.Error()), err
			}

			client := http.Client{}
			for _, k := range headers.Keys() {
				key := k.(starlark.String).GoString()
				value, _, _ := headers.Get(k)
				// log.Printf("Constructing header key=%s and value=%s", key, value.(starlark.String).GoString())
				req.Header.Add(key, value.(starlark.String).GoString())
			}
			res, err := client.Do(req)
			if err != nil {
				return starlark.String(err.Error()), err
			}

			defer res.Body.Close()

			resBytes, err := io.ReadAll(res.Body)
			if err != nil {
				return starlark.String(err.Error()), err
			}

			return starlark.String(string(resBytes)), nil
		}),

		"getEnv": starlark.NewBuiltin("getEnv", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			value := os.Getenv(args[0].(starlark.String).GoString())
			return starlark.String(value), nil
		}),
	}

	globals, err := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, "sf.star", sf, predeclared)
	if err != nil {
		log.Printf("Unable to execute the Starlark function. Err is %s", err)
		return "", err
	}

	sfState := starlark.Tuple{}

	for _, p := range r.definition.Parameters {
		pv, exists := params[p.Name]
		if !exists {
			sfState = append(sfState, starlark.NoneType(' '))
			continue
		}
		var starlarkParam starlark.Value

		switch p.Type {
		case "string":
			starlarkParam = starlark.String(pv.(string))
		case "integer":
			starlarkParam = starlark.MakeInt(pv.(int))
		case "boolean":
			starlarkParam = starlark.Bool(pv.(bool))
		}

		sfState = append(sfState, starlarkParam)

	}

	fnName := strings.ToLower(r.definition.Name)
	fn, exists := globals[fnName]
	if !exists {
		return "", fmt.Errorf("expected function %s not found", fnName)
	}

	res, err := starlark.Call(thread, fn, sfState, nil)
	if err != nil {
		log.Printf("Error executing the Starlark function %s", err)
		return "", err
	}

	log.Printf("RESPONSE FROM STARLARK FUNCTION CALL IS %s", res)

	return res.(starlark.String).GoString(), nil
}
