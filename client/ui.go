package main

import (
	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
	"github.com/omesh-barhate/coderpad/client/editor"
)

// TUI is built using termbox-go.
// termbox allows us to set any content to individual cells, and hence, the basic building block of the editor is a "cell".

// UI creates a new editor view and runs the main loop.
func UI(connection *websocket.Conn) error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	ed = editor.NewEditor()
	ed.SetSize(termbox.Size())
	ed.Draw()

	err = mainLoop(connection)
	if err != nil {
		return err
	}

	return nil
}

// mainLoop is the main update loop for the UI.
func mainLoop(connection *websocket.Conn) error {
	// termboxChannel is used for sending and receiving termbox events.
	termboxChannel := getTermboxChan()

	// messageChannel is used for sending and receiving messages.
	messageChannel := getMsgChan(connection)

	for {
		select {
		case event := <-termboxChannel:
			err := handleTermboxEvent(event, connection)
			if err != nil {
				return err
			}
		case message := <-messageChannel:
			handleMsg(message, connection)
		}
	}
}
