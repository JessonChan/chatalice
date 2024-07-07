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
	case "getChats":
		list := store.GetChatList()
		resp := []map[string]any{}
		for _, chat := range list {
			messages := store.GetMessageList(chat.ChatID)
			messageResp := []map[string]any{}
			for _, message := range messages {
				if message.Content == "" {
					continue
				}
				messageResp = append(messageResp, map[string]any{
					"id":     message.ID,
					"isUser": message.Role == "user",
					"text":   message.Content,
				})
			}
			resp = append(resp, map[string]any{
				"id":       chat.ChatID,
				"title":    chat.Title,
				"messages": messageResp,
			})
		}
		return resp
	case "hello":
		msg := store.Message{}
		json.Unmarshal([]byte(args), &msg)
		messages := store.GetMessageList(msg.ChatID)
		model := store.GetModelByID(msg.ModelID)
		chat := store.GetChatByChatID(msg.ChatID)
		if chat.ID == 0 {
			// 新建一个Chat
			chat = store.Chat{
				Title:  "Untitled",
				ChatID: msg.ChatID,
			}
			store.InsertChat(&chat)
		}
		if chat.Title == "Untitled" {
			go func() {
				title := llm.Title(model, msg.Content)
				store.UpdateChatTitleByChatID(msg.ChatID, title)
				runtime.EventsEmit(a.ctx, "updateChatTitle", toJSON(map[string]any{
					"id":    msg.ChatID,
					"title": title,
				}))
			}()
		}
		// TODO a lot of checks
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
