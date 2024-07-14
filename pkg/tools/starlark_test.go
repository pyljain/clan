package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	std := StarlarkTool{
		Name:        "TestTool",
		Description: "A Starlark tool for testing",
		Parameters: []StarlarkToolParameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Test string",
			},
			{
				Name:        "location",
				Type:        "string",
				Description: "Test string",
			},
		},
		Function: `def print_hello(name, location):
			print("Hello " + name + location)`,
	}

	tool := NewStarlarkHandler(&std)
	schema := tool.Schema()

	assert.Equal(t, "TestTool", schema.Name)
	assert.Equal(t, "A Starlark tool for testing", schema.Description)

	props, exists := schema.Schema["properties"]
	assert.True(t, exists)

	properties := props.(map[string]interface{})
	_, nameExists := properties["name"]
	assert.True(t, nameExists)

	_, locationExists := properties["location"]
	assert.True(t, locationExists)
}

func TestExecute(t *testing.T) {
	std := StarlarkTool{
		Name:        "PrintHello",
		Description: "A Starlark tool for testing",
		Parameters: []StarlarkToolParameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Test string",
			},
			{
				Name:        "location",
				Type:        "string",
				Description: "Test string",
			},
		},
		Function: `def printhello(name, location):
			print("Hello " + name + location)
			return "Hello " + name + location`,
	}

	tool := NewStarlarkHandler(&std)
	output, err := tool.Execute(map[string]interface{}{
		"name":     "Dodgy",
		"location": "London",
	})

	assert.NoError(t, err)
	assert.Equal(t, output, "Hello DodgyLondon")
}

func TestExecuteErrorWhenFunctionNameIncorrect(t *testing.T) {
	std := StarlarkTool{
		Name:        "Hello",
		Description: "A Starlark tool for testing",
		Parameters: []StarlarkToolParameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Test string",
			},
			{
				Name:        "location",
				Type:        "string",
				Description: "Test string",
			},
		},
		Function: `def hi(name, location):
			return "Hello " + name + location`,
	}

	tool := NewStarlarkHandler(&std)
	_, err := tool.Execute(map[string]interface{}{
		"name":     "Dodgy",
		"location": "London",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected function hello not found")
}

func TestExecuteWithHTTPGet(t *testing.T) {
	std := StarlarkTool{
		Name:        "HTTPCall",
		Description: "A Starlark tool for testing",
		Parameters: []StarlarkToolParameter{
			{
				Name:        "term",
				Type:        "string",
				Description: "actors",
			},
		},
		Function: `def httpcall(term):
			result = get('http://google.com?q=' + term)
			return result`,
	}

	tool := NewStarlarkHandler(&std)
	output, err := tool.Execute(map[string]interface{}{
		"term": "Dodgy",
	})

	assert.NoError(t, err)
	assert.NotEqual(t, output, "")
}

func TestExecuteWithGetEnv(t *testing.T) {
	std := StarlarkTool{
		Name:        "getEnvTest",
		Description: "A Starlark tool for retrieving an environment variable",
		Parameters:  []StarlarkToolParameter{},
		Function: `def getenvtest():
			result = getEnv("TEST_VAR")
			return result`,
	}

	tool := NewStarlarkHandler(&std)
	output, err := tool.Execute(map[string]interface{}{})

	assert.NoError(t, err)
	assert.Equal(t, "1", output)
}

func TestExecuteWithHTTPPost(t *testing.T) {
	std := StarlarkTool{
		Name:        "HTTPCall",
		Description: "A Starlark tool for testing",
		Parameters: []StarlarkToolParameter{
			{
				Name:        "term",
				Type:        "string",
				Description: "actors",
			},
		},
		Function: `def httpcall(term):
		result = post('https://google.serper.dev/search', {
			"X-API-KEY": getEnv("SERPER_KEY"),
			"Content-Type": "application/json"
		}, '{"q": "' + term + '"}')
		return result`,
	}

	tool := NewStarlarkHandler(&std)
	output, err := tool.Execute(map[string]interface{}{
		"term": "Dodgy",
	})

	assert.NoError(t, err)
	assert.NotEqual(t, output, "")
}
