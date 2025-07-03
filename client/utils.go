package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/omesh-barhate/coderpad/crdt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

type Arguments struct {
	ServerAddress string
	UseSecure     bool
	RequireLogin  bool
	FilePath      string
	EnableDebug   bool
}

func parseFlags() Arguments {
	serverAddress := flag.String("server", "localhost:8080", "The network address of the server")
	useSecure := flag.Bool("secure", false, "Enable a secure WebSocket connection (wss://)")
	enableDebug := flag.Bool("debug", false, "Enable debugging mode to show more verbose logs")
	requireLogin := flag.Bool("login", false, "Enable the login prompt for the server")
	filePath := flag.String("file", "", "The file to load the coderpad content from")

	flag.Parse()

	return Arguments{
		ServerAddress: *serverAddress,
		UseSecure:     *useSecure,
		EnableDebug:   *enableDebug,
		RequireLogin:  *requireLogin,
		FilePath:      *filePath,
	}
}

func createConnection(arguments Arguments) (*websocket.Conn, *http.Response, error) {
	var wsURL url.URL
	if arguments.UseSecure {
		wsURL = url.URL{Scheme: "wss", Host: arguments.ServerAddress, Path: "/"}
	} else {
		wsURL = url.URL{Scheme: "ws", Host: arguments.ServerAddress, Path: "/"}
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: 2 * time.Minute,
	}
	return dialer.Dial(wsURL.String(), nil)
}

func ensureDirectoryExists(directoryPath string) (bool, error) {
	if _, err := os.Stat(directoryPath); err == nil {
		return true, nil
	}
	err := os.Mkdir(directoryPath, 0700)
	if err != nil {
		return false, err
	}
	return true, nil
}

func setupLogger(logInstance *logrus.Logger) (*os.File, *os.File, error) {
	logFilePath := "coderpad.log"
	debugLogFilePath := "coderpad-debug.log"

	homeDirExists := true
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDirExists = false
	}

	applicationDir := filepath.Join(homeDir, ".coderpad")
	dirExists, err := ensureDirectoryExists(applicationDir)
	if err != nil {
		return nil, nil, err
	}

	if dirExists && homeDirExists {
		logFilePath = filepath.Join(applicationDir, "coderpad.log")
		debugLogFilePath = filepath.Join(applicationDir, "coderpad-debug.log")
	}

	logFileHandle, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Logger error, exiting: %s", err)
		return nil, nil, err
	}

	debugLogFileHandle, err := os.OpenFile(debugLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Logger error, exiting: %s", err)
		return nil, nil, err
	}

	logInstance.SetOutput(io.Discard)
	logInstance.SetFormatter(&logrus.JSONFormatter{})
	logInstance.AddHook(&writer.Hook{
		Writer: logFileHandle,
		LogLevels: []logrus.Level{
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})
	logInstance.AddHook(&writer.Hook{
		Writer: debugLogFileHandle,
		LogLevels: []logrus.Level{
			logrus.TraceLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
		},
	})

	return logFileHandle, debugLogFileHandle, nil
}

func closeLogFiles(logFileHandle, debugLogFileHandle *os.File) {
	if err := logFileHandle.Close(); err != nil {
		fmt.Printf("Failed to close log file: %s", err)
		return
	}
	if err := debugLogFileHandle.Close(); err != nil {
		fmt.Printf("Failed to close debug log file: %s", err)
		return
	}
}

func printDocument(document crdt.Document) {
	if arguments.EnableDebug {
		logger.Infof("---DOCUMENT STATE---")
		for i, character := range document.Characters {
			logger.Infof("index: %v  value: %s  ID: %v  PrevID: %v  NextID: %v  ", i, character.Value, character.ID, character.PrevID, character.NextID)
		}
	}
}
