package store

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

func getDb(initDb ...bool) *gorm.DB {
	configPath, err := os.UserConfigDir()
	if err != nil {
		// TODO 更好的错误处理
		panic(err)
	}
	dbFilePath := filepath.Join(configPath, "chat.db")
	db, err := gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	if len(initDb) > 0 && initDb[0] {
		db.AutoMigrate(&Model{})
	}
	return db
}

type Model struct {
	gorm.Model
	Name    string `json:"name"`
	API_KEY string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

type Chat struct {
	gorm.Model
	Title string `json:"title"`
}

type Message struct {
	gorm.Model
	Content string `json:"content"`
	ChatID  uint   `json:"chat_id"`
	ModelID uint   `json:"model_id"`
	Role    string `json:"role"`
}

func GetModelList() []Model {
	db := getDb()
	var models []Model
	db.Find(&models)
	return models
}

func InsertModel(model Model) {
	db := getDb(true)
	db.Create(&model)
}

func DeleteModelByID(id uint) {
	db := getDb()
	m := &Model{}
	m.ID = id
	db.Delete(m)
}

func GetChatList() []Chat {
	db := getDb()
	var chats []Chat
	db.Find(&chats)
	return chats
}

func InsertChat(chat Chat) {
	db := getDb(true)
	db.Create(&chat)
}

func DeleteChatByID(id uint) {
	db := getDb()
	c := &Chat{}
	c.ID = id
	db.Delete(c)
}

func GetMessageList(chatID uint) []Message {
	db := getDb()
	var messages []Message
	db.Where("chat_id = ?", chatID).Find(&messages)
	return messages
}

func InsertMessage(message Message) uint {
	db := getDb(true)
	db.Create(&message)
	return message.ID
}

func DeleteMessageByID(id uint) {
	db := getDb()
	m := &Message{}
	m.ID = id
	db.Delete(m)
}
func UpdateMessageContentByID(id uint, content string) {
	db := getDb()
	db.Model(&Message{}).Where("id = ?", id).Update("content", content)
}
