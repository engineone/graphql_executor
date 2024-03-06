package graphql_executor

import (
	"context"
	"encoding/json"

	"github.com/asaskevich/govalidator"
	http_executor "github.com/engineone/http_executor/executor"
	"github.com/engineone/types"
	"github.com/palantir/stacktrace"
)

type GraphQLExecutor struct {
	inputRules  map[string]interface{}
	outputRules map[string]interface{}
	*http_executor.HttpExecutor
}

// NewGraphQLExecutor creates a new GraphQLExecutor
func NewGraphQLExecutor() *GraphQLExecutor {
	return &GraphQLExecutor{
		inputRules: map[string]interface{}{
			"url":     "required,url",
			"headers": "required,dictionary",
			"body":    "required,dictionary",
		},
		outputRules: map[string]interface{}{
			"headers": "required,dictionary",
			"body":    "required,dictionary",
		},
		HttpExecutor: http_executor.NewHttpExecutor(),
	}
}

func (e *GraphQLExecutor) New() *GraphQLExecutor {
	return NewGraphQLExecutor()
}

func (e *GraphQLExecutor) ID() string {
	return "graphql"
}

func (e *GraphQLExecutor) Name() string {
	return "GraphQL"
}

func (e *GraphQLExecutor) Description() string {
	return "GraphQL executor to make http requests to a given url with the given method and headers."
}

func (e *GraphQLExecutor) InputRules() map[string]interface{} {
	return e.inputRules
}

func (e *GraphQLExecutor) OutputRules() map[string]interface{} {
	return e.outputRules
}

func (e *GraphQLExecutor) Validate(ctx context.Context, task *types.Task, tasks []*types.Task) error {
	if task.Input == nil {
		return stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input is required")
	}

	input, ok := task.Input.(map[string]interface{})
	if !ok {
		return stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input must be an object")
	}

	_, err := govalidator.ValidateMap(input, e.inputRules)
	return stacktrace.PropagateWithCode(err, types.ErrInvalidTask, "Input validation failed")
}

func (e *GraphQLExecutor) Execute(ctx context.Context, task *types.Task, wf []*types.Task) (interface{}, error) {
	input, ok := task.Input.(map[string]interface{})
	if !ok {
		return nil, stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input must be an object")
	}
	input["method"] = "POST"

	// Set the content type to application/json
	headers, ok := input["headers"].(map[string]interface{})
	if !ok {
		headers = make(map[string]interface{})
	}
	headers["Content-Type"] = "application/json"
	input["headers"] = headers

	// Put the headers and input back into the task
	task.Input = input

	out, err := e.HttpExecutor.Execute(ctx, task, wf)
	if err != nil {
		return out, stacktrace.Propagate(err, "failed to execute graphql executor")
	}

	output := make(map[string]interface{})
	if err := json.Unmarshal(out.(map[string]interface{})["body"].([]byte), &output); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal output")
	}

	return output, nil
}
