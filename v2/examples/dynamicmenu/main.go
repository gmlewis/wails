package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	app.makeMenu()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "dynamicmenu",
		Width:  1024,
		Height: 768,
		Menu:   app.appMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *App) makeMenu() {
	a.appMenu = menu.NewMenu()

	a.fileMenu = a.appMenu.AddSubmenu("File - A")

	a.langMenu = a.fileMenu.AddSubmenu("Language - A")
	a.langMenu.AddRadio("A", true, nil, a.changeLanguage)
	a.langMenu.AddRadio("B", false, nil, a.changeLanguage)
	a.langMenu.AddRadio("C", false, nil, a.changeLanguage)

	a.fileMenu.AddSeparator()
	a.fileMenu.AddText("Quit - A", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})
}

func (a *App) changeLanguage(data *menu.CallbackData) {
	if a.language == data.MenuItem.Label {
		return
	}
	a.language = data.MenuItem.Label

	a.appMenu.Items[0].Label = fmt.Sprintf("File - %v", a.language)
	a.fileMenu.Items[0].Label = fmt.Sprintf("Language - %v", a.language)
	a.fileMenu.Items[2].Label = fmt.Sprintf("Quit - %v", a.language)

	runtime.MenuSetApplicationMenu(a.ctx, a.appMenu) // crashes on Linux
	// runtime.MenuUpdateApplicationMenu(a.ctx) // does nothing on Linux
}
