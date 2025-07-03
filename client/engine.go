package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
	"github.com/omesh-barhate/coderpad/commons"
	"github.com/omesh-barhate/coderpad/crdt"
	"github.com/sirupsen/logrus"
)

func handleTermboxEvent(event termbox.Event, connection *websocket.Conn) error {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEsc, termbox.KeyCtrlC:
			return errors.New("coderpad: exiting")
		case termbox.KeyCtrlS:
			if fileName == "" {
				fileName = "coderpad-content.txt"
			}
			err := crdt.Save(fileName, &document)
			if err != nil {
				ed.StatusMsg = "Failed to save to " + fileName
				logrus.Errorf("failed to save to %s", fileName)
				ed.SetStatusBar()
				return err
			}
			ed.StatusMsg = "Saved document to " + fileName
			ed.SetStatusBar()
		case termbox.KeyCtrlL:
			if fileName != "" {
				logger.Log(logrus.InfoLevel, "LOADING DOCUMENT")
				newDocument, err := crdt.Load(fileName)
				ed.StatusMsg = "Loading " + fileName
				ed.SetStatusBar()
				if err != nil {
					ed.StatusMsg = "Failed to load " + fileName
					logrus.Errorf("failed to load file %s", fileName)
					ed.SetStatusBar()
					return err
				}
				document = newDocument
				ed.SetX(0)
				ed.SetText(crdt.Content(document))
				logger.Log(logrus.InfoLevel, "SENDING DOCUMENT")
				message := commons.Message{MessageType: commons.DocSyncMessage, Document: document}
				_ = connection.WriteJSON(&message)
			} else {
				ed.StatusMsg = "No file to load!"
				ed.SetStatusBar()
			}
		case termbox.KeyArrowLeft, termbox.KeyCtrlB:
			ed.MoveCursor(-1, 0)
		case termbox.KeyArrowRight, termbox.KeyCtrlF:
			ed.MoveCursor(1, 0)
		case termbox.KeyArrowUp, termbox.KeyCtrlP:
			ed.MoveCursor(0, -1)
		case termbox.KeyArrowDown, termbox.KeyCtrlN:
			ed.MoveCursor(0, 1)
		case termbox.KeyHome:
			ed.SetX(0)
		case termbox.KeyEnd:
			ed.SetX(len(ed.Text))
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			performOperation(OperationDelete, event, connection)
		case termbox.KeyDelete:
			performOperation(OperationDelete, event, connection)
		case termbox.KeyTab:
			for i := 0; i < 4; i++ {
				event.Ch = ' '
				performOperation(OperationInsert, event, connection)
			}
		case termbox.KeyEnter:
			event.Ch = '\n'
			performOperation(OperationInsert, event, connection)
		case termbox.KeySpace:
			event.Ch = ' '
			performOperation(OperationInsert, event, connection)
		default:
			if event.Ch != 0 {
				performOperation(OperationInsert, event, connection)
			}
		}
	}
	ed.Draw()
	return nil
}

const (
	OperationInsert = iota
	OperationDelete
)

func performOperation(operationType int, event termbox.Event, connection *websocket.Conn) {
	character := string(event.Ch)
	var message commons.Message
	switch operationType {
	case OperationInsert:
		logger.Infof("LOCAL INSERT: %s at cursor position %v\n", character, ed.Cursor)
		runes := []rune(character)
		ed.AddRune(runes[0])
		text, err := document.Insert(ed.Cursor, character)
		if err != nil {
			ed.SetText(text)
			logger.Errorf("CRDT error: %v\n", err)
		}
		ed.SetText(text)
		message = commons.Message{MessageType: "operation", Operation: commons.Operation{OperationType: "insert", Position: ed.Cursor, Value: character}}
	case OperationDelete:
		logger.Infof("LOCAL DELETE: cursor position %v\n", ed.Cursor)
		if ed.Cursor-1 < 0 {
			ed.Cursor = 0
		}
		text := document.Delete(ed.Cursor)
		ed.SetText(text)
		message = commons.Message{MessageType: "operation", Operation: commons.Operation{OperationType: "delete", Position: ed.Cursor}}
		ed.MoveCursor(-1, 0)
	}
	err := connection.WriteJSON(message)
	if err != nil {
		ed.StatusMsg = "lost connection!"
		ed.SetStatusBar()
	}
}

func getTermboxChan() chan termbox.Event {
	termboxChannel := make(chan termbox.Event)
	go func() {
		for {
			termboxChannel <- termbox.PollEvent()
		}
	}()
	return termboxChannel
}

func handleMsg(message commons.Message, connection *websocket.Conn) {
	switch message.MessageType {
	case commons.DocSyncMessage:
		logger.Infof("DOCSYNC RECEIVED, updating local document %+v\n", message.Document)
		document = message.Document
	case commons.DocReqMessage:
		logger.Infof("DOCREQ RECEIVED, sending local document to %v\n", message.ClientID)
		response := commons.Message{MessageType: commons.DocSyncMessage, Document: document, ClientID: message.ClientID}
		_ = connection.WriteJSON(&response)
	case commons.SiteIDMessage:
		siteID, err := strconv.Atoi(message.Text)
		if err != nil {
			logger.Errorf("failed to set siteID, err: %v\n", err)
		}
		crdt.SiteID = siteID
		logger.Infof("SITE ID %v, INTENDED SITE ID: %v", crdt.SiteID, siteID)
	case commons.JoinMessage:
		ed.StatusMsg = fmt.Sprintf("%s has joined the session!", message.Username)
		ed.SetStatusBar()
	default:
		switch message.Operation.OperationType {
		case "insert":
			_, err := document.Insert(message.Operation.Position, message.Operation.Value)
			if err != nil {
				logger.Errorf("failed to insert, err: %v\n", err)
			}
			logger.Infof("REMOTE INSERT: %s at position %v\n", message.Operation.Value, message.Operation.Position)
		case "delete":
			_ = document.Delete(message.Operation.Position)
			logger.Infof("REMOTE DELETE: position %v\n", message.Operation.Position)
		}
	}
	printDocument(document)
	ed.SetText(crdt.Content(document))
	ed.Draw()
}

func getMsgChan(connection *websocket.Conn) chan commons.Message {
	messageChannel := make(chan commons.Message)
	go func() {
		for {
			var message commons.Message
			err := connection.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Errorf("websocket error: %v", err)
				}
				break
			}
			logger.Infof("message received: %+v\n", message)
			messageChannel <- message
		}
	}()
	return messageChannel
}
