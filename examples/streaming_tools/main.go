package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	openairesponses "github.com/gosticks/openai-responses-api-go"
)

// WeatherParams represents the parameters for the weather function
type WeatherParams struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

// getWeather is a mock function to get the weather
func getWeather(location, unit string) string {
	// In a real application, this would call a weather API
	return fmt.Sprintf("The weather in %s is sunny and 25 degrees %s.", location, unit)
}

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(1)
	}

	// Create a new client
	client := openairesponses.NewClient(apiKey)

	// Define the weather function parameters schema
	weatherParamsSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The city and state, e.g. San Francisco, CA",
			},
			"unit": map[string]interface{}{
				"type": "string",
				"enum": []string{"celsius", "fahrenheit"},
			},
		},
		"required": []string{"location"},
	}

	// Create a new streaming response with a function tool
	fmt.Println("Creating streaming response with function tool...")
	stream, err := client.Responses.CreateStream(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.DeveloperInputMessage("You are a helpful assistant with access to weather information."),
				openairesponses.UserInputMessage("What's the weather like in San Francisco?"),
			},
			Tools: []openairesponses.ResponseTool{
				openairesponses.NewFunctionTool(
					"get_weather",
					"Get the current weather in a given location",
					weatherParamsSchema,
				),
			},
			// Set instructions for the model
			Instructions: "You are a weather assistant. Always use the get_weather function to retrieve weather information.",
		},
	)
	if err != nil {
		fmt.Printf("Error creating streaming response: %v\n", err)
		os.Exit(1)
	}
	defer stream.Close()

	// Create an accumulator to accumulate the streaming responses
	accumulator := &openairesponses.ResponsesStreamAccumulator{}

	// Process the streaming response
	fmt.Println("Streaming response:")
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("\nStream closed")
			break
		}
		if err != nil {
			fmt.Printf("Error receiving chunk: %v\n", err)
			os.Exit(1)
		}

		// Add the chunk to the accumulator
		accumulator.AddChunk(chunk)

		// Print the chunk content if available
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fmt.Print(choice.Delta.Content)
			}

			// Print tool calls if available
			if len(choice.Delta.ToolCalls) > 0 {
				for _, toolCall := range choice.Delta.ToolCalls {
					fmt.Printf("\nTool call received - Function: %s\n", toolCall.Function.Name)
					if toolCall.Function.Arguments != "" {
						fmt.Printf("Arguments: %s\n", toolCall.Function.Arguments)
					}
				}
			}
		}
	}

	// Convert the accumulator to a response
	resp := accumulator.ToResponse()

	// Check if the model wants to call a function
	if len(resp.Choices) > 0 && len(resp.Choices[0].ToolCalls) > 0 {
		fmt.Println("\n\nProcessing tool calls...")

		// Process each tool call
		for _, toolCall := range resp.Choices[0].ToolCalls {
			if toolCall.Function.Name == "get_weather" {
				// Parse the function arguments
				var params WeatherParams
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					fmt.Printf("Error parsing function arguments: %v\n", err)
					continue
				}

				// Call the function
				unit := params.Unit
				if unit == "" {
					unit = "celsius"
				}
				result := getWeather(params.Location, unit)
				fmt.Printf("\nFunction result: %s\n", result)

				// Create a new streaming response with the function result using previous_response_id
				fmt.Println("\nCreating follow-up streaming response with function result...")
				followUpStream, err := client.Responses.CreateStream(
					context.Background(),
					openairesponses.ResponseRequest{
						Model: "gpt-4o",
						Input: []openairesponses.ResponseInputMessage{
							// Only need to provide the new user message and tool result
							openairesponses.SystemInputMessage(fmt.Sprintf("Function get_weather returned: %s", result)),
						},
						// Use the previous response ID to continue the conversation
						PreviousResponseID: resp.ID,
					},
				)
				if err != nil {
					fmt.Printf("Error creating follow-up streaming response: %v\n", err)
					os.Exit(1)
				}
				defer followUpStream.Close()

				// Create a new accumulator for the follow-up response
				followUpAccumulator := &openairesponses.ResponsesStreamAccumulator{}

				// Process the follow-up streaming response
				fmt.Println("Follow-up streaming response:")
				for {
					chunk, err := followUpStream.Recv()
					if err == io.EOF {
						fmt.Println("\nFollow-up stream closed")
						break
					}
					if err != nil {
						fmt.Printf("Error receiving follow-up chunk: %v\n", err)
						os.Exit(1)
					}

					// Add the chunk to the accumulator
					followUpAccumulator.AddChunk(chunk)

					// Print the chunk content if available
					for _, choice := range chunk.Choices {
						if choice.Delta.Content != "" {
							fmt.Print(choice.Delta.Content)
						}
					}
				}

				// Convert the follow-up accumulator to a response
				followUpResp := followUpAccumulator.ToResponse()

				// Ask a follow-up question using previous_response_id
				fmt.Println("\n\nAsking a follow-up question...")
				followUpQuestion, err := client.Responses.CreateStream(
					context.Background(),
					openairesponses.ResponseRequest{
						Model: "gpt-4o",
						Input: []openairesponses.ResponseInputMessage{
							openairesponses.UserInputMessage("How does that compare to the weather in New York?"),
						},
						// Use the previous response ID to continue the conversation
						PreviousResponseID: followUpResp.ID,
						// Tools are still available from the previous response
						Tools: []openairesponses.ResponseTool{
							openairesponses.NewFunctionTool(
								"get_weather",
								"Get the current weather in a given location",
								weatherParamsSchema,
							),
						},
					},
				)
				if err != nil {
					fmt.Printf("Error creating follow-up question: %v\n", err)
					os.Exit(1)
				}
				defer followUpQuestion.Close()

				// Create a new accumulator for the follow-up question
				questionAccumulator := &openairesponses.ResponsesStreamAccumulator{}

				// Process the follow-up question
				fmt.Println("Follow-up question response:")
				for {
					chunk, err := followUpQuestion.Recv()
					if err == io.EOF {
						fmt.Println("\nFollow-up question stream closed")
						break
					}
					if err != nil {
						fmt.Printf("Error receiving follow-up question chunk: %v\n", err)
						os.Exit(1)
					}

					// Add the chunk to the accumulator
					questionAccumulator.AddChunk(chunk)

					// Print the chunk content if available
					for _, choice := range chunk.Choices {
						if choice.Delta.Content != "" {
							fmt.Print(choice.Delta.Content)
						}

						// Print tool calls if available
						if len(choice.Delta.ToolCalls) > 0 {
							for _, toolCall := range choice.Delta.ToolCalls {
								fmt.Printf("\nTool call received - Function: %s\n", toolCall.Function.Name)
								if toolCall.Function.Arguments != "" {
									fmt.Printf("Arguments: %s\n", toolCall.Function.Arguments)
								}
							}
						}
					}
				}

				// Convert the question accumulator to a response
				questionResp := questionAccumulator.ToResponse()

				// Process any tool calls from the follow-up question
				if len(questionResp.Choices) > 0 && len(questionResp.Choices[0].ToolCalls) > 0 {
					fmt.Println("\n\nProcessing follow-up tool calls...")

					// Process each tool call
					for _, toolCall := range questionResp.Choices[0].ToolCalls {
						if toolCall.Function.Name == "get_weather" {
							// Parse the function arguments
							var params WeatherParams
							if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
								fmt.Printf("Error parsing function arguments: %v\n", err)
								continue
							}

							// Call the function
							unit := params.Unit
							if unit == "" {
								unit = "celsius"
							}
							result := getWeather(params.Location, unit)
							fmt.Printf("\nFunction result: %s\n", result)

							// Create a final response with the function result
							fmt.Println("\nCreating final response with function result...")
							finalStream, err := client.Responses.CreateStream(
								context.Background(),
								openairesponses.ResponseRequest{
									Model: "gpt-4o",
									Input: []openairesponses.ResponseInputMessage{
										openairesponses.SystemInputMessage(fmt.Sprintf("Function get_weather returned: %s", result)),
									},
									// Use the previous response ID to continue the conversation
									PreviousResponseID: questionResp.ID,
								},
							)
							if err != nil {
								fmt.Printf("Error creating final response: %v\n", err)
								os.Exit(1)
							}
							defer finalStream.Close()

							// Create a new accumulator for the final response
							finalAccumulator := &openairesponses.ResponsesStreamAccumulator{}

							// Process the final response
							fmt.Println("Final response:")
							for {
								chunk, err := finalStream.Recv()
								if err == io.EOF {
									fmt.Println("\nFinal stream closed")
									break
								}
								if err != nil {
									fmt.Printf("Error receiving final chunk: %v\n", err)
									os.Exit(1)
								}

								// Add the chunk to the accumulator
								finalAccumulator.AddChunk(chunk)

								// Print the chunk content if available
								for _, choice := range chunk.Choices {
									if choice.Delta.Content != "" {
										fmt.Print(choice.Delta.Content)
									}
								}
							}

							// Convert the final accumulator to a response
							finalResp := finalAccumulator.ToResponse()

							// Print usage information if available
							if finalResp.Usage != nil {
								fmt.Printf("\n\nFinal usage information:\n")
								fmt.Printf("  Prompt tokens: %d\n", finalResp.Usage.PromptTokens)
								fmt.Printf("  Completion tokens: %d\n", finalResp.Usage.CompletionTokens)
								fmt.Printf("  Total tokens: %d\n", finalResp.Usage.TotalTokens)
							}
						}
					}
				}

				// Print usage information if available
				if questionResp.Usage != nil {
					fmt.Printf("\nFollow-up question usage information:\n")
					fmt.Printf("  Prompt tokens: %d\n", questionResp.Usage.PromptTokens)
					fmt.Printf("  Completion tokens: %d\n", questionResp.Usage.CompletionTokens)
					fmt.Printf("  Total tokens: %d\n", questionResp.Usage.TotalTokens)
				}

				// Print usage information if available
				if followUpResp.Usage != nil {
					fmt.Printf("\nFollow-up usage information:\n")
					fmt.Printf("  Prompt tokens: %d\n", followUpResp.Usage.PromptTokens)
					fmt.Printf("  Completion tokens: %d\n", followUpResp.Usage.CompletionTokens)
					fmt.Printf("  Total tokens: %d\n", followUpResp.Usage.TotalTokens)
				}
			}
		}
	}

	// Print usage information if available
	if resp.Usage != nil {
		fmt.Printf("\nInitial usage information:\n")
		fmt.Printf("  Prompt tokens: %d\n", resp.Usage.PromptTokens)
		fmt.Printf("  Completion tokens: %d\n", resp.Usage.CompletionTokens)
		fmt.Printf("  Total tokens: %d\n", resp.Usage.TotalTokens)
	}
}