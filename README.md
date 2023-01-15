## server_clipboard

A server which saves my clipboard (in memory), so I can share it between my devices.

I use [termux](https://termux.com/) on my phone to communicate with my server (with `server_clipboard <c|p>` (copy/paste))

On other devices I don't have a terminal on, this has a web interface at `/`:

<img src="https://github.com/seanbreckenridge/server_clipboard/blob/main/frontend/demo.png" alt="screencap of server html page">

### Run

Run `server_clipboard server` on a remote server somewhere, update your `~/.bashrc`/`~/.zshenv` to include a password/remote address:

```
export CLIPBOARD_PASSWORD='i8nCzZnSY4hlHwUF9Ny15vqtPjfezpMHKZll0030Gn1p17Uiw7'
export CLIPBOARD_ADDRESS='http://mywebsite.com/clipboard'
```

```
NAME:
   server_clipboard - share clipboard between devices using a server

USAGE:
   server_clipboard [global options] command [command options] [arguments...]

COMMANDS:
   server, s  start server
   copy, c    copy to server clipboard
   paste, p   paste from server clipboard
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value, -p value  port to listen on (default: 5025) [$CLIPBOARD_PORT]
   --password value        password to use [$CLIPBOARD_PASSWORD]
   --server_address value  server address to connect to (default: "localhost:5025") [$CLIPBOARD_ADDRESS]
   --help, -h              show help (default: false)
```

This automatically detects which operating system you're on and uses the corresponding clipboard commands, see [`clipboard.go`](clipboard.go). If this cant, set the `CLIPBOARD_PASTE_COMMAND` and `CLIPBOARD_COPY_COMMAND` environment variables

### Install

Install `golang` (requires `1.18`+)

You can clone and run `go build`, or:

```
go install -v "github.com/seanbreckenridge/server_clipboard/cmd/server_clipboard@latest"
```

which downloads, builds and puts the binary on your `$GOBIN`.
