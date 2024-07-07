package main

import (
	"chatalice/llm"
	"chatalice/store"
	"context"
	"encoding/json"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Call(fn string, args string) string {
	return toJSON(a.call(fn, args))
}
func (a *App) call(fn string, args string) any {
	fmt.Println("func called", fn, args)
	switch fn {
	case "getModelList":
		return store.GetModelList()
	case "insertModel":
		fmt.Println("insertModel here")
		m := store.Model{}
		err := json.Unmarshal([]byte(args), &m)
		fmt.Println("insertModel here", err)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("insertModel: %v\n", m)
		store.InsertModel(m)
	case "":
	case "hello":
		msg := store.Message{}
		json.Unmarshal([]byte(args), &msg)
		messages := store.GetMessageList(msg.ChatID)
		model := store.GetModelByID(msg.ModelID)
		fmt.Println("model:", model)
		if msg.ChatID == 0 {
			// msg.ChatID = store.InsertChat(store.Chat{
			// SystemPrompt: "you are a bot",
			// Title:        "New Chat",
			// })
		}
		msg.Role = "user"
		store.InsertMessage(msg)
		answerID := store.InsertMessage(store.Message{
			ChatID:  msg.ChatID,
			Role:    "assistant",
			Content: "",
		})
		go llm.Stream(model, messages, msg.Content, func(chuckText string) {
			fmt.Println("callback", answerID)
			bs, _ := json.Marshal(map[string]any{
				"message_id": answerID,
				"text":       chuckText,
			})
			store.UpdateMessageContentByID(answerID, chuckText)
			runtime.EventsEmit(a.ctx, "appendMessage", string(bs))
		})
		return map[string]any{
			"message_id": answerID,
			"text":       "",
		}
	}
	return ""
}

func toJSON(obj any) string {
	if obj == nil {
		return ""
	}
	switch obj := obj.(type) {
	case string:
		return obj
	}
	bs, _ := json.Marshal(obj)
	return string(bs)
}
