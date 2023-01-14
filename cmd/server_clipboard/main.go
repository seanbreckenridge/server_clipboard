package main

import (
	"fmt"
	"github.com/seanbreckenridge/server_clipboard"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
)

func main() {
	app := &cli.App{
		Name:  "server_clipboard",
		Usage: "share clipboard between devices using a server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   5025,
				Usage:   "port to listen on",
				EnvVars: []string{"CLIPBOARD_PORT"},
			},
			&cli.StringFlag{
				Name:     "password",
				Value:    "",
				Usage:    "password to use",
				Required: true,
				EnvVars:  []string{"CLIPBOARD_PASSWORD"},
			},
			&cli.StringFlag{
				Name:     "server_address",
				Value:    "localhost:5025",
				Usage:    "server address to connect to",
				Required: true,
				EnvVars:  []string{"CLIPBOARD_ADDRESS"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "start server",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						Value:   false,
						Usage:   "enable debug logging",
					},
				},
				Action: func(c *cli.Context) error {
					return server_clipboard.Server(c.String("password"), c.Int("port"), c.Bool("debug"))
				},
			},
			{
				Name:    "copy",
				Aliases: []string{"c"},
				Usage:   "copy to server clipboard",
				Action: func(c *cli.Context) error {
					text, err := server_clipboard.Copy(c.String("password"), c.String("server_address"), server_clipboard.FetchClipboard(c.String("clipboard")))
					if err != nil {
						return err
					}
					if strings.TrimSpace(text) != "" {
						fmt.Fprintln(os.Stderr, text)
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "clipboard",
						EnvVars:  []string{"CLIPBOARD_CONTENTS"},
						Usage:    "clipboard data to upload to server",
						Required: false,
					},
				},
			},
			{
				Name:    "paste",
				Aliases: []string{"p"},
				Usage:   "paste from server clipboard",
				Action: func(c *cli.Context) error {
					text, err := server_clipboard.Paste(c.String("password"), c.String("server_address"))
					if err != nil {
						return err
					}

					if strings.TrimSpace(text) != "" {
						err := server_clipboard.SetClipboard(text)
						if err != nil {
							// if we have text, print text regardless of if there was an error
							fmt.Println(text)
							return err
						} else {
							fmt.Fprintln(os.Stderr, "pasted into local clipboard")
						}
					} else {
						fmt.Fprintln(os.Stderr, "server returned empty clipboard")
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
