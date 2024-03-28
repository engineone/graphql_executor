package graphql_executor_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	graphql_executor "github.com/engineone/graphql_executor/executor"
	"github.com/engineone/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GraphQLExecutor", func() {
	var (
		executor *graphql_executor.GraphQLExecutor
		task     *types.Task
	)

	BeforeEach(func() {
		executor = graphql_executor.NewGraphQLExecutor()
		task = &types.Task{
			Input: map[string]interface{}{
				"url":     "http://localhost:8080",
				"headers": map[string]interface{}{"Content-Type": "application/json"},
				"query":   "{ hello }",
			},
		}
	})

	Describe("Execute", func() {
		Context("with valid input", func() {
			It("should return the correct output", func() {
				// Create a test server that returns a fixed response
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"data": {"hello": "world"}}`))
				}))
				defer server.Close()

				// Update the task URL to point to the test server
				input := task.Input.(map[string]interface{})
				input["url"] = server.URL
				task.Input = input

				output, err := executor.Execute(context.Background(), task, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(map[string]interface{}{
					"data": map[string]interface{}{
						"hello": "world",
					},
				}))
			})
		})
	})
})
