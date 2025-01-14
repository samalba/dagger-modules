// A generated module for DockerToDagger functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/docker-to-dagger/internal/dagger"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type DockerToDagger struct{}

func (m *DockerToDagger) systemPrompt(ctx context.Context) (string, error) {
	message := `You are generating Dagger modules from Dockerfiles.

- Dagger is a container engine that allows building containers from native Code.
- Anything that can be done in a Dockerfile can be expressed with the Dagger API.
- You will be provided with a Dockerfile as input, and you will return the code of a Go function that does exactly the same behavior using the Dagger API.

Here is code in Golang that shows several examples of Dockerfiles and there equivalent in the Dagger API.

The functions are numbered such as Example1, Example2, Example<number>, etc...

Each function is an implementation of the Dockerfile in comment in the Docstring of the function.
Each line of the function maps to each line in the example Dockerfile, there are extra explanation in comments.
`

	exampleFile, err := dag.CurrentModule().Source().File("examples/main.go").Contents(ctx)
	if err != nil {
		return "", err
	}

	message = fmt.Sprintf("%s\n```go\n%s\n```\n", message, exampleFile)

	return message, nil
}

func (m *DockerToDagger) generateDaggerModule(moduleCode string) *dagger.Directory {
	return dag.CurrentModule().
		Source().
		Directory("examples").
		WithNewFile("main.go", moduleCode)
}

// Convert a Dockerfile into a new Dagger module
func (m *DockerToDagger) FromDockerfile(ctx context.Context, anthropicApiKey *dagger.Secret, dockerfile *dagger.File) (*dagger.Directory, error) {
	dockerfileContent, err := dockerfile.Contents(ctx)
	if err != nil {
		return nil, err
	}

	userPrompt := "Convert the following Dockerfile between the code block to a Dagger function that produces the same container:\n\n"
	userPrompt += fmt.Sprintf("```\n%s\n```\n\n", dockerfileContent)
	userPrompt += `For returning the code, follow the additional instructions:
	- Return a complete Dagger module containing a function named "Build" that implements the Dockerfile.
	- The Example code provided is a valid Dagger module, you can follow this format.
	- Like the functions provided as examples, add a generic comment on the first line of the function Docstring, followed by the original Dockerfile.
	- The import path of the module will always be the same as the one provided in the example.
	- The main Struct in the Dagger module should be named "MyModule".
	- Double check the import paths of the generated code, for example: only include the "context" import path if you use a context in the function.
	- Do not add explanations around the code block, return only the Go code (without surrounding backquotes), so it can be compiled directly.`

	apiKey, err := anthropicApiKey.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	systemPrompt, err := m.systemPrompt(ctx)
	if err != nil {
		return nil, err
	}

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
		MaxTokens: anthropic.F(int64(2048)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewAssistantMessage(anthropic.NewTextBlock(systemPrompt)),
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		}),
	})
	if err != nil {
		panic(err.Error())
	}

	moduleCode := ""
	for _, m := range message.Content {
		if m.Type != anthropic.ContentBlockTypeText {
			continue
		}
		moduleCode = m.Text
	}

	if moduleCode == "" {
		return nil, fmt.Errorf("no code found in response: %+v", message)
	}

	return m.generateDaggerModule(moduleCode), nil
}
