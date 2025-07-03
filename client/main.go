package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/omesh-barhate/coderpad/client/editor"
	"github.com/omesh-barhate/coderpad/commons"
	"github.com/omesh-barhate/coderpad/crdt"
	"github.com/sirupsen/logrus"
)

var (
	document  = crdt.New()
	logger    = logrus.New()
	ed        = editor.NewEditor()
	fileName  string
	arguments Arguments
)

func main() {
	arguments = parseFlags()
	scanner := bufio.NewScanner(os.Stdin)
	userName := randomdata.SillyName()
	if arguments.RequireLogin {
		fmt.Print("Enter your name: ")
		scanner.Scan()
		userName = scanner.Text()
	}
	connection, _, err := createConnection(arguments)
	if err != nil {
		fmt.Printf("Connection error, exiting: %s\n", err)
		return
	}
	defer connection.Close()
	joinMessage := commons.Message{Username: userName, Text: "has joined the session.", MessageType: commons.JoinMessage}
	_ = connection.WriteJSON(joinMessage)
	logFile, debugLogFile, err := setupLogger(logger)
	if err != nil {
		fmt.Printf("Failed to setup logger, exiting: %s\n", err)
		return
	}
	defer closeLogFiles(logFile, debugLogFile)
	document = crdt.New()
	if arguments.FilePath != "" {
		if document, err = crdt.Load(arguments.FilePath); err != nil {
			fmt.Printf("failed to load document: %s\n", err)
			return
		}
	}
	err = UI(connection)
	if err != nil {
		if strings.HasPrefix(err.Error(), "coderpad") {
			fmt.Println("exiting session.")
			return
		}
		fmt.Printf("TUI error, exiting: %s\n", err)
		return
	}
}
