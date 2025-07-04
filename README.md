# coderpad

A playful, hackable collaborative text editor toy written in Go.

---

## 🚀 Quick Start

1. **Start the server** (in one terminal):
   ```sh
   go run server/main.go
   # or
   go run main.go
   ```
   _Default address: `:8080`_

2. **Start a client** (in another terminal):
   ```sh
   go run client/*.go
   ```
   _Spin up multiple clients for collaborative editing!_

---

## ✨ Features

- Super lightweight (~4MB binary)
- Easy to run (single binary or `go run`)
- Export/import document content
- Built for hacking and learning
- Collaborative editing (CRDT-backed)
- Not for production—just for fun!

## ⌨️ Keybindings

| Action                | Key(s)                        |
|---------------------- |:-----------------------------:|
| Exit                  | `Esc`, `Ctrl+C`               |
| Save document         | `Ctrl+S`                      |
| Load document         | `Ctrl+L`                      |
| Move cursor left      | `Left`, `Ctrl+B`              |
| Move cursor right     | `Right`, `Ctrl+F`             |
| Move cursor up        | `Up`, `Ctrl+P`                |
| Move cursor down      | `Down`, `Ctrl+N`              |
| Move to line start    | `Home`                        |
| Move to line end      | `End`                         |
| Delete character      | `Backspace`, `Delete`         |

---

## 🛠 Usage

### Server
```
Usage of coderpad-server:
  -addr string
        Server's network address (default ":8080")
```

### Client
```
Usage of coderpad:
  -debug         Enable verbose debug logs
  -file string   Load coderpad content from file
  -login         Enable login prompt
  -secure        Use secure WebSocket (wss://)
  -server string Server address (default "localhost:8080")
```

---

## 🧠 How does it work?

- Each client maintains a CRDT-backed local document state.
- The server:
  - Manages client connections
  - Broadcasts operations to all clients
- Clients:
  - Connect and send operations to the server
  - Render the document in a TUI
  - Handle key events and dispatch changes

---

**coderpad** is a toy project. Fork it, break it, hack it, and have fun!
