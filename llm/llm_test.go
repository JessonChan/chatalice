package llm

import (
	"chatalice/store"
	"os"
	"testing"
)

var model = store.Model{
	ModelName: "gpt-4o-mini",
	Key:       os.Getenv("OPENAI_API_KEY"),
	BaseURL:   os.Getenv("OPENAI_API_BASE_URL"),
}

func TestTitle(t *testing.T) {
	title := Title(model, "hello world")
	t.Log(title)
}

func TestWithPic(t *testing.T) {

}
