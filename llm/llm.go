package llm

import (
	"chatalice/store"
	"encoding/json"

	"context"
	"errors"
	"fmt"
	"io"
	"strings"

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
			Role: "system",
			// TODO 标题应该也允许用户自己设定和选择
			Content: "你是一个非常善于起标题的机器人，可以根据用户的一个输入,想到可能的输出，生成一个不超过16个字的标题，并且标题由一个emoji和空格开始，如：❤️ 这是一个标题格式。你的回复务必只需要给定这个标题本身，不需要其它任何的多余信息或符号",
		},
		{
			Role:    "user",
			Content: userInput,
		},
	}
	resp, err := c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:     model.ModelName,
		Messages:  messages,
		MaxTokens: 16,
	})
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	fmt.Printf("%+v", resp)
	title = resp.Choices[0].Message.Content
	return
}

type UserInput struct {
	Content string
	Images  []string
	Files   []string
}

func (userInput UserInput) toLLMMessage() openai.ChatCompletionMessage {
	fmt.Println("userInput.Content:", userInput.Content, userInput.Images, len(userInput.Images))
	if len(userInput.Images) > 0 {
		message := openai.ChatCompletionMessage{
			Role:         "user",
			MultiContent: []openai.ChatMessagePart{},
		}
		if userInput.Content != "" {
			message.MultiContent = append(message.MultiContent, openai.ChatMessagePart{
				Text: userInput.Content,
				Type: "text",
			})
		}
		for _, image := range userInput.Images {
			message.MultiContent = append(message.MultiContent, openai.ChatMessagePart{
				ImageURL: &openai.ChatMessageImageURL{
					URL:    image,
					Detail: openai.ImageURLDetailAuto,
				},
				Type: "image_url",
			})
		}
		return message
	}
	return openai.ChatCompletionMessage{
		Role:    "user",
		Content: userInput.Content,
	}
}

func Stream(model store.Model, chat store.Chat, msgHistory []store.Message, userInput UserInput, callback func(string)) {
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
		for _, msg := range msgHistory {
			if preRole == msg.Role {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    map[string]string{"user": "assistant", "assistant": "user"}[preRole],
					Content: fillContent(""),
				})
			}
			if msg.Images != "" {
				message := openai.ChatCompletionMessage{
					Role:         msg.Role,
					MultiContent: []openai.ChatMessagePart{},
				}
				if msg.Content != "" {
					message.MultiContent = append(message.MultiContent, openai.ChatMessagePart{
						Text: fillContent(msg.Content),
						Type: "text",
					})
				}
				images := strings.Split(msg.Images, "&")
				for _, image := range images {
					if image == "" {
						continue
					}
					message.MultiContent = append(message.MultiContent, openai.ChatMessagePart{
						ImageURL: &openai.ChatMessageImageURL{
							URL:    image,
							Detail: openai.ImageURLDetailAuto,
						},
						Type: "image_url",
					})
				}
				messages = append(messages, message)
			} else {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    msg.Role,
					Content: fillContent(msg.Content),
				})
			}
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
	if model.SystemPromptEnabled {
		messages = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: chat.SystemPrompt,
			},
		}, messages...)
	}
	messages = append(messages, userInput.toLLMMessage())
	dbs, _ := json.Marshal(messages)
	fmt.Println(string(dbs))
	fmt.Println("Streaming...", msgHistory, messages)
	stream(c, model, chat.MaxOutputTokens, messages, callback)
}

func stream(c *openai.Client, model store.Model, maxOutputTokens int, messages []openai.ChatCompletionMessage, callback func(string)) {
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		// Model: "alibaba/Qwen2-7B-Instruct",
		// Model: openai.GPT3Dot5Turbo,
		// Model: "claude-3-5-sonnet-20240620",
		MaxTokens: maxOutputTokens,
		Model:     model.ModelName,
		Messages:  messages,
		Stream:    model.StreamResponseEnabled,
	}
	if model.StreamResponseEnabled {
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
	} else {
		resp, err := c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model:     model.ModelName,
			Messages:  messages,
			MaxTokens: 16,
		})
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}
		fmt.Printf("%+v", resp)
		callback(resp.Choices[0].Message.Content)
	}
}
