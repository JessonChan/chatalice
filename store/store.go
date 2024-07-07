package store

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	getDb(true)
}

var sqlDB *gorm.DB

func getDb(initDb ...bool) *gorm.DB {
	if sqlDB != nil {
		return sqlDB
	}
	configPath, err := os.UserConfigDir()
	if err != nil {
		// TODO 更好的错误处理
		panic(err)
	}
	// TODO dbFilePath := filepath.Join(filepath.Join(configPath, "ChatAlice"), "chat.db")
	dbFilePath := filepath.Join(configPath, "chat.db")
	fmt.Println("db file path:", dbFilePath)
	db, err := gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Info, // 设置日志级别为 "debug"
			}),
	})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	if len(initDb) > 0 && initDb[0] {
		db.AutoMigrate(&Model{})
		db.AutoMigrate(&Chat{})
		db.AutoMigrate(&Message{})
	}
	sqlDB = db
	return db
}

type Model struct {
	gorm.Model
	Name      string `json:"name"`
	ModelName string `json:"model"`
	Key       string `json:"key"`
	BaseURL   string `json:"baseUrl"`
}

type Chat struct {
	gorm.Model
	Title        string `json:"title"`
	ChatID       uint   `json:"chatId"`
	SystemPrompt string `json:"systemPrompt"`
}

type Message struct {
	gorm.Model
	Content string `json:"content"`
	ChatID  uint   `json:"chatId"`
	ModelID uint   `json:"modelId"`
	Role    string `json:"role"`
}

func GetModelList() []Model {
	db := getDb()
	var models []Model
	db.Find(&models)
	return models
}

func InsertModel(model Model) {
	db := getDb()
	err := db.Create(&model)
	if err != nil {
		fmt.Println(err)
	}
}

func DeleteModelByID(id uint) {
	db := getDb()
	m := &Model{}
	m.ID = id
	db.Delete(m)
}

func GetModelByID(id uint) Model {
	db := getDb()
	var model Model
	db.First(&model, id)
	return model
}

func GetChatList() []Chat {
	db := getDb()
	var chats []Chat
	db.Order("id desc").Find(&chats)
	return chats
}

func GetChatByChatID(id uint) Chat {
	db := getDb()
	var chat Chat
	db.Where("chat_id = ?", id).Find(&chat)
	return chat
}

func UpdateChatTitleByChatID(id uint, title string) {
	db := getDb()
	db.Model(&Chat{}).Where("chat_id = ?", id).Update("title", title)
}

func InsertChat(chat *Chat) uint {
	db := getDb()
	db.Create(chat)
	return chat.ID
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
	db.Where("chat_id = ?", chatID).Order("id asc").Find(&messages)
	return messages
}

func InsertMessage(message Message) uint {
	db := getDb()
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
	fmt.Println("update message content", id, content)
	db := getDb()
	er := db.Model(&Message{}).Where("id = ?", id).Update("content", gorm.Expr("content || ?", content))
	if er.Error != nil {
		fmt.Sprintf("update error %v", er.Error)
	}
}
