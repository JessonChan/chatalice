package main

import (
	"chatalice/store"
	"context"
	"encoding/json"
	"fmt"
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
	fmt.Println(fn, args)
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
