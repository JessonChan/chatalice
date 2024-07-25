package store

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var sqlDB *gorm.DB

func getDb() *gorm.DB {
	if sqlDB != nil {
		return sqlDB
	}
	configPath, err := os.UserConfigDir()
	if err != nil {
		// TODO 更好的错误处理
		panic(err)
	}
	// 附加程序特定的目录路径
	appName := "ChatAlice"
	appConfigDir := filepath.Join(configPath, appName)

	// 创建目录
	err = os.MkdirAll(appConfigDir, 0755)
	if err != nil {
		fmt.Println("无法创建程序配置目录:", err)
		panic(err)
	}
	// TODO dbFilePath := filepath.Join(filepath.Join(configPath, "ChatAlice"), "chat.db")
	dbFilePath := filepath.Join(appConfigDir, "chat.db")
	fmt.Println("db file path:", dbFilePath)
	fi, err := os.Stat(dbFilePath)
	dbInit := err != nil || fi == nil || fi.Size() == 0
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
	if dbInit {
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
	Title              string `json:"title"`
	ChatID             uint   `json:"chatId"` // is a unique id,created by client
	ModelID            uint   `json:"modelId"`
	ConversationRounds int    `json:"conversationRounds"`
	MaxInputTokens     int    `json:"maxInputTokens"`
	MaxOutputTokens    int    `json:"maxOutputTokens"`
	SystemPrompt       string `json:"systemPrompt"`
}

var DefaulChatConfig = Chat{
	Title:              "Untitled",
	ConversationRounds: 3,
	MaxInputTokens:     4096,
	MaxOutputTokens:    4096,
	SystemPrompt:       "You are a helpful assistant.",
}

func NewChat(chatId, modelId uint) Chat {
	return Chat{
		Title:              "Untitled",
		ChatID:             chatId,
		ModelID:            modelId,
		ConversationRounds: 3,
		MaxInputTokens:     4096,
		MaxOutputTokens:    4096,
		SystemPrompt:       "You are a helpful assistant.",
	}
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

func GetChatListByUpdatedAt(updateAt time.Time) []Chat {
	db := getDb()
	var chats []Chat
	db.Order("updated_at desc").Where("updated_at <?", updateAt).Limit(20).Find(&chats)
	return chats
}

func GetChatByChatID(id uint) Chat {
	db := getDb()
	var chat Chat
	db.Where("chat_id = ?", id).Find(&chat)
	return chat
}

func SaveChatSetting(chat *Chat) {
	db := getDb()
	db.Save(chat)
}

func UpdateChatTitleByChatID(chatId uint, title string) {
	db := getDb()
	db.Model(&Chat{}).Where("chat_id = ?", chatId).Update("title", title)
}
func UpdateChatModelIDByChatID(chatId, modelId uint) {
	db := getDb()
	db.Model(&Chat{}).Where("chat_id = ?", chatId).Update("model_id", modelId)
}

func UpdateChatLatestTime(chatId uint) {
	db := getDb()
	db.Model(&Chat{}).Where("chat_id = ?", chatId).Updates(map[string]interface{}{})
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
