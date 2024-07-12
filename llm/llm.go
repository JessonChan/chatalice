package llm

import (
	"chatalice/store"

	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

func Title(model store.Model, userInput string) (title string) {
	defer func() {
		if title == "" {
			title = "Untitled"
		}
	}()
	clietConfig := openai.DefaultConfig(model.Key)
	clietConfig.BaseURL = model.BaseURL
	c := openai.NewClientWithConfig(clietConfig)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "你是一个非常善于起标题的机器，可以根据用户的一个输入,想到可能的输出，生成一个不超过16个字的标题,你的回复务必只需要给定这个标题本身，不需要其它任何的多余信息或符号",
		},
		{
			Role:    "user",
			Content: userInput,
		},
	}
	resp, err := c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    model.ModelName,
		Messages: messages,
	})
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	title = resp.Choices[0].Message.Content
	return
}

func Stream(model store.Model, chat store.Chat, msgHistory []store.Message, userInput string, callback func(string)) {
	clietConfig := openai.DefaultConfig(model.Key)
	clietConfig.BaseURL = model.BaseURL
	c := openai.NewClientWithConfig(clietConfig)
	messages := []openai.ChatCompletionMessage{}
	fillContent := func(text string) string {
		if text != "" {
			return text
		}
		return "continue"
	}
	if len(msgHistory) > 0 {
		preRole := msgHistory[0].Role
		if preRole != "user" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    "user",
				Content: fillContent(""),
			})
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msgHistory[0].Role,
			Content: fillContent(msgHistory[0].Content),
		})
		for _, msg := range msgHistory[1:] {
			if preRole == msg.Role {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    map[string]string{"user": "assistant", "assistant": "user"}[preRole],
					Content: fillContent(""),
				})
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: fillContent(msg.Content),
			})
			preRole = msg.Role
		}
	}
	if len(messages) >= chat.ConversationRounds*2 {
		messages = messages[len(messages)-chat.ConversationRounds*2:]
	}
	if len(messages) > 0 {
		if messages[len(messages)-1].Role != "assistant" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: fillContent(""),
			})
		}
	}
	messages = append([]openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: chat.SystemPrompt,
		},
	}, messages...)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: userInput,
	})
	fmt.Println("Streaming...", msgHistory, messages)
	// TODO chat.MaxInputTokens setting here
	stream(c, model.ModelName, chat.MaxOutputTokens, messages, callback)
}
func stream(c *openai.Client, model string, maxOutputTokens int, messages []openai.ChatCompletionMessage, callback func(string)) {
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		// Model: "alibaba/Qwen2-7B-Instruct",
		// Model: openai.GPT3Dot5Turbo,
		// Model: "claude-3-5-sonnet-20240620",
		MaxTokens: maxOutputTokens,
		Model:     model,
		Messages:  messages,
		Stream:    true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		fmt.Println(response, err)
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			// return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		callback(response.Choices[0].Delta.Content)
	}
}
