package server_clipboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func Server(password string, port int, debug bool) error {
	// keep track of server state
	var clipboard Clipboard
	var lock = sync.RWMutex{}
	clipboard = Clipboard{Text: "", Timestamp: time.Now()}

	// start server
	http.HandleFunc("/copy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		// check password
		if r.Header.Get("password") != password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid password"))
			return
		}

		// read body
		body, err := ioutil.ReadAll(r.Body)
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
		log.Printf("copying '%s' to clipboard\n", input.Text)
		clipboard = Clipboard{Text: input.Text, Timestamp: time.Now()}

		w.Header().Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated remote clipboard"))
	})

	http.HandleFunc("/paste", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		// check password
		if r.Header.Get("password") != password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid password"))
			return
		}

		// write clipboard to response, use rlock to allow concurrent connections
		lock.RLock()
		defer lock.RUnlock()

		log.Printf("fetching '%s' from clipboard\n", clipboard.Text)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(clipboard.Text))

	})

	fmt.Fprintf(os.Stderr, "listening on port %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
