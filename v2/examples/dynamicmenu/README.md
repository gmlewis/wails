# README

This example demonstrates an issue on Linux with the following commands:

* runtime.MenuSetApplicationMenu(a.ctx, a.appMenu)  // crashes on Linux
  To duplicate the crash, first select File => Language => B
  Note that the menus did not change.
  Then select File => Language => C
  This causes a crash in linux.handleMenuItemClick in the file
  `internal/frontend/desktop/linux/gtk.go` (line 51) because the gtkWidget (item) is `nil`.
* runtime.MenuUpdateApplicationMenu(a.ctx)  // ignored on Linux
  No matter what language is chosen, the call is always ignored.

## About

This is the official Wails Vanilla template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.
