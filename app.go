package main

import (
	"chatalice/llm"
	"chatalice/store"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		fmt.Println("getChats", args)
		lastSeen, _ := strconv.Atoi(args)
		lastUpdateAt := time.Unix(int64(lastSeen), 0)
		// list := store.GetChatList()
		list := store.GetChatListByUpdatedAt(lastUpdateAt)
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
				"id":                 chat.ChatID,
				"title":              chat.Title,
				"messages":           messageResp,
				"modelId":            chat.ModelID,
				"conversationRounds": chat.ConversationRounds,
				"maxInputTokens":     chat.MaxInputTokens,
				"maxOutputTokens":    chat.MaxOutputTokens,
				"systemPrompt":       chat.SystemPrompt,
				"updatedAt":          chat.UpdatedAt.Unix(),
			})
		}
		return resp
	case "updateChatSetting":
		// TODO why this is so hard and confusing to understand
		setting := struct {
			ChatId             uint   `json:"chatId"`
			ModelId            uint   `json:"modelId"`
			ConversationRounds int    `json:"conversationRounds"`
			MaxInputTokens     int    `json:"maxInputTokens"`
			MaxOutputTokens    int    `json:"maxOutputTokens"`
			SystemPrompt       string `json:"systemPrompt"`
		}{}
		json.Unmarshal([]byte(args), &setting)
		chat := store.GetChatByChatID(setting.ChatId)
		if chat.ID == 0 {
			chat = store.NewChat(setting.ChatId, setting.ModelId)
		}
		chat.ConversationRounds = setting.ConversationRounds
		chat.MaxInputTokens = setting.MaxInputTokens
		chat.MaxOutputTokens = setting.MaxOutputTokens
		chat.ModelID = setting.ModelId
		chat.SystemPrompt = setting.SystemPrompt
		store.SaveChatSetting(&chat)
	case "sendMessage":
		msg := store.Message{}
		json.Unmarshal([]byte(args), &msg)
		messages := store.GetMessageList(msg.ChatID)
		model := store.GetModelByID(msg.ModelID)
		chat := store.GetChatByChatID(msg.ChatID)
		if chat.ID == 0 {
			// 新建一个Chat
			chat = store.NewChat(msg.ChatID, msg.ModelID)
			store.InsertChat(&chat)
		} else {
			store.UpdateChatLatestTime(msg.ChatID)
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
		if chat.ModelID != msg.ModelID {
			store.UpdateChatModelIDByChatID(msg.ChatID, msg.ModelID)
		}
		// TODO a lot of checks
		msg.Role = "user"
		store.InsertMessage(msg)
		answerID := store.InsertMessage(store.Message{
			ChatID:  msg.ChatID,
			Images:  msg.Images,
			Role:    "assistant",
			Content: "",
		})
		images := []string{}
		if len(msg.Images) > 0 {
			images = strings.Split(msg.Images, "&")
		}
		fullMessage := ``
		go llm.Stream(model, chat, messages, llm.UserInput{Content: msg.Content, Images: images}, func(chuckText string) {
			if chuckText == "" {
				return
			}
			fmt.Println("callback", answerID)
			fullMessage += chuckText
			bs, _ := json.Marshal(map[string]any{
				"message_id": answerID,
				"text":       chuckText,
			})
			store.UpdateMessageContentByID(answerID, chuckText)
			// runtime.EventsEmit(a.ctx, "updateMessage", string(bs))
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
