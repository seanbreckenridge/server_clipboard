package server_clipboard

import (
	"errors"
	"fmt"
	"github.com/seanbreckenridge/on_machine"
	"log"
	"os"
	"os/exec"
	"strings"
)

func commandOutput(command string) string {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error running %s: %s", command, err))
	}
	return string(out)
}

func FetchClipboard(clipboard string) string {
	if clipboard != "" {
		return clipboard
	}
	if cmd := os.Getenv("CLIPBOARD_COPY_COMMAND"); cmd != "" {
		return commandOutput(cmd)
	}

	if on_machine.OnTermux() {
		return commandOutput("termux-clipboard-get")
	} else {
		switch on_machine.GetOS() {
		case "linux":
			return commandOutput("xclip -o -selection clipboard")
		case "mac":
			return commandOutput("pbpaste")
		case "windows":
			return commandOutput("powershell.exe -Command Get-Clipboard")
		default:
			log.Fatal(fmt.Printf("Unsupported OS: %s. Set the CLIPBOARD_COPY_COMMAND environment variable to a command which prints your clipboard", on_machine.GetOS()))
		}
	}
	return ""
}

func commandWithStdin(command string, stdin string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader(stdin)
	err := cmd.Run()
	return err
}

func SetClipboard(clipboard string) error {
	if cmd := os.Getenv("CLIPBOARD_PASTE_COMMAND"); cmd != "" {
		return commandWithStdin(cmd, clipboard)
	}

	if on_machine.OnTermux() {
		return commandWithStdin("termux-clipboard-set", clipboard)
	}

	switch on_machine.GetOS() {
	case "linux":
		return commandWithStdin("xclip -i -selection clipboard", clipboard)
	case "mac":
		return commandWithStdin("pbcopy", clipboard)
	case "windows":
		return commandWithStdin("powershell.exe -Command Set-Clipboard", clipboard)
	default:
		return errors.New(fmt.Sprintf("Unsupported OS: %s. Set the CLIPBOARD_PASTE_COMMAND environment variable to a command which sets your clipboard", on_machine.GetOS()))
	}
}
