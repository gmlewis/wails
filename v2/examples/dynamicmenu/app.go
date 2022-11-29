package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/menu"
)

// App struct
type App struct {
	ctx context.Context

	language string

	appMenu  *menu.Menu
	fileMenu *menu.Menu
	langMenu *menu.Menu
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		language: "A",
	}
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
