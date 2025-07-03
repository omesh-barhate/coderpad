
# coderpad

A playful, hackable collaborative text editor toy written in Go.


## Features

- Super lightweight (~4MB)
- Easy to run (single binary, or just `go run`!)
- Export/import your document content
- Built for hacking and learning
- Not for production—just for fun!

## Keybindings

| Action         | Key |
|--------------|:-----:|
| Exit |  `Esc`, `Ctrl+C` |
| Save to document |  `Ctrl+S` |
| Load from document |  `Ctrl+L` |
| Move cursor left |  `Left arrow key`, `Ctrl+B` |
| Move cursor right |  `Right arrow key`, `Ctrl+F` |
| Move cursor up |  `Up arrow key`, `Ctrl+P` |
| Move cursor down |  `Down arrow key`, `Ctrl+N` |
| Move cursor to start |  `Home` |
| Move cursor to end |  `End` |
| Delete characters |  `Backspace`, `Delete` |

## How to play

Start the server (in one terminal):

```
./coderpad-server
```

```
Usage of coderpad-server:
  -addr string
        Server's network address (default ":8080")
```

Then start a client (in another terminal):

```
./coderpad
```

```
Usage of coderpad:
  -debug
        Enable debugging mode to show more verbose logs
  -file string
        The file to load the coderpad content from
  -login
        Enable the login prompt for the server
  -secure
        Enable a secure WebSocket connection (wss://)
  -server string
        The network address of the server (default "localhost:8080")
```

Example play:

- Connect to a server: `coderpad -server coderpad.test`
- Enable login prompt: `coderpad -server coderpad.test -login`
- Specify a file to save to/load from: `coderpad -server coderpad.test -file example.txt`
- Enable debugging mode: `coderpad -server coderpad.test -debug`

### Local setup (for hackers)

To start the server:

```
go run server/main.go
```

To start the client:

```
go run client/*.go
```

(Spin up at least 2 clients for the full collaborative toy experience! It also works with a single client.)

## How does it work?

- Each client has a CRDT-backed local state (document).
- The CRDT is a sequence of characters with some attributes.
- The server:
  - establishes connections with clients
  - keeps a list of active connections
  - broadcasts operations from one client to all others
- Clients connect and send operations to the server.
- The TUI:
  - Renders document content
  - Handles key events
  - Generates and dispatches payloads on key presses

---

**coderpad** is a toy project. Fork it, break it, hack it, and have fun!
