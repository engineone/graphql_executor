package graphql_executor

import (
	"context"
	"encoding/json"

	http_executor "github.com/engineone/http_executor/executor"
	"github.com/engineone/types"
	"github.com/engineone/utils"
	validate "github.com/go-playground/validator/v10"
	"github.com/palantir/stacktrace"
)

type Input struct {
	URL       string                 `json:"url" valid:"required,url"`
	Headers   map[string]string      `json:"headers" valid:"dictionary"`
	Query     string                 `json:"query" valid:"required"`
	Variables map[string]interface{} `json:"variables" valid:"dictionary"`
}

type Output struct {
	Headers map[string]string      `json:"headers" valid:"required,dictionary"`
	Body    map[string]interface{} `json:"body" valid:"required,dictionary"`
}

type GraphQLExecutor struct {
	*http_executor.HttpExecutor
	inputCache *Input
}

// NewGraphQLExecutor creates a new GraphQLExecutor
func NewGraphQLExecutor() *GraphQLExecutor {
	return &GraphQLExecutor{
		HttpExecutor: http_executor.NewHttpExecutor(),
	}
}

func (e *GraphQLExecutor) New() types.Executor {
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
	return utils.ExtractValidationRules(&Input{})
}

func (e *GraphQLExecutor) OutputRules() map[string]interface{} {
	return utils.ExtractValidationRules(&Output{})
}

func (e *GraphQLExecutor) convertInput(input interface{}) (*Input, error) {
	if e.inputCache != nil {
		return e.inputCache, nil
	}

	e.inputCache = &Input{}
	if err := utils.ConvertToStruct(input, e.inputCache); err != nil {
		return nil, stacktrace.PropagateWithCode(err, types.ErrInvalidTask, "Error converting input to struct")
	}
	return e.inputCache, nil
}

func (e *GraphQLExecutor) Validate(ctx context.Context, task *types.Task, tasks []*types.Task) error {
	if task.Input == nil {
		return stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input is required")
	}

	var err error
	e.inputCache, err = e.convertInput(task.Input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to convert input")
	}

	v := validate.New()
	if err := v.Struct(e.inputCache); err != nil {
		return stacktrace.PropagateWithCode(err, types.ErrInvalidTask, "Input validation failed")
	}
	return nil
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

	// Use query and variables from the input to put together the body
	body := map[string]interface{}{
		"query":     input["query"],
		"variables": input["variables"],
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal body")
	}

	input["body"] = string(bodyBytes)

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
