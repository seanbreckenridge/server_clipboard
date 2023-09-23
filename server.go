package server_clipboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Clipboard struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// json input for copy
type CopyInput struct {
	Text string `json:"text"`
}

//go:embed frontend/dist/index.html
var index embed.FS

func Server(password string, port int, debug bool, clearAfter int) error {
	// keep track of server state
	lock := sync.RWMutex{}
	clipboard := Clipboard{Text: "", Timestamp: time.Now()}

	indexData, err := index.ReadFile("frontend/dist/index.html")
	if err != nil {
		return err
	}

	if clearAfter > 0 {
		log.Printf("server will clear clipboard %d seconds after last copy\n", clearAfter)
	}

	var clearAt *time.Time = nil

	// start a goroutine to run the clear loop
	// it can just wait 1/5th of the clearAfter time, and check if the time has expired
	// if it has, clear the clipboard
	go func() {
		if clearAfter <= 0 {
			return
		}
		for {
			if debug {
				log.Printf("clearAfter: sleeping for %d seconds\n", clearAfter/5)
			}
			// sleep for 1/5th of the clearAfter time
			time.Sleep(time.Duration(clearAfter/5) * time.Second)
			if clearAt == nil {
				continue
			}
			lock.Lock()
			if time.Now().After(*clearAt) {
				if debug {
					log.Printf("clearAfter: has been %d seconds since clipboard was last set, clearing clipboard\n", clearAfter)
				}

				clipboard = Clipboard{Text: "", Timestamp: time.Now()}
				clearAt = nil
			}
			lock.Unlock()
		}
	}()

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			// write index to response
			w.Header().Set("Content-Type", "text/html")
			w.Write(indexData)
		})

	// start server
	http.HandleFunc("/copy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// allow cross origin requests
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// check password
		if r.Header.Get("password") != password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid password"))
			return
		}

		// read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error reading body"))
			return
		}
		// to json
		var input CopyInput
		if err := json.Unmarshal(body, &input); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error parsing JSON body"))
			return
		}

		// unlock for writing
		lock.Lock()
		defer lock.Unlock()
		if debug {
			log.Printf("copying '%s' to clipboard\n", input.Text)
		}
		clipboard = Clipboard{Text: input.Text, Timestamp: time.Now()}

		w.Header().Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated remote clipboard"))

		// set clearAt time
		if clearAfter > 0 {
			ca := time.Now().Add(time.Duration(clearAfter) * time.Second)
			clearAt = &ca
			if debug {
				log.Printf("clearAfter: set clearAt to %s\n", clearAt.Format(time.RFC3339))
			}
		}
	})

	http.HandleFunc("/paste", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// allow cross origin requests
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// check password
		if r.Header.Get("password") != password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid password"))
			return
		}

		// write clipboard to response, use rlock to allow concurrent connections
		lock.RLock()
		defer lock.RUnlock()

		if debug {
			log.Printf("fetching '%s' from clipboard\n", clipboard.Text)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("X-Clipboard-Timestamp", clipboard.Timestamp.Format(time.RFC822Z))
		w.Write([]byte(clipboard.Text))

	})

	fmt.Fprintf(os.Stderr, "listening on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
