package server_clipboard

import (
	"embed"
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

//go:embed frontend/dist/index.html
var index embed.FS

func Server(password string, port int, debug bool, clearAfter int) error {
	// keep track of server state
	lock := sync.RWMutex{}
	clipboard := Clipboard{Text: "", Timestamp: time.Now()}
	clearLoopRunning := false

	indexData, err := index.ReadFile("frontend/dist/index.html")
	if err != nil {
		return err
	}

	if clearAfter > 0 {
		log.Printf("server will clear clipboard %d seconds after last copy\n", clearAfter)
	}

	// code to clear clipboard after x seconds
	// send a message to 'clearer' when a copy is made (if clearAfter > 0)
	clearer := make(chan bool)
	go func() {
		var clearAt *time.Time
		clearAt = nil
		for {
			select {
			// code blocks here until a message is received on 'clearer'
			case <-clearer:
				// start timer
				if clearAt != nil {
					if debug {
						log.Printf("clearAfter: reset timer to clear clipboard after %d seconds\n", clearAfter)
					}
				} else {
					if debug {
						log.Printf("clearAfter: started timer to clear clipboard after %d seconds\n", clearAfter)
					}
				}
				ca := time.Now().Add(time.Duration(clearAfter) * time.Second)
				clearAt = &ca
				if clearLoopRunning {
					if debug {
						log.Printf("clearAfter: timer already running, skipping starting a new one\n")
					}
					continue
				}

				// start a goroutuine to wait for the timer
				// this should not sleep exactly the time, but rather
				// check some fraction of the clearAfter to see if the
				// time has expired, since its possible that the timer
				// gets reset to be a higher clearAt time before the timer expires
				go func() {
					// sleep for a fraction of the clearAfter time, but at least 1 second
					// and at most 30 seconds
					sleepSecs := clearAfter / 10
					if sleepSecs < 1 {
						sleepSecs = 1
					} else if sleepSecs > 30 {
						sleepSecs = 30
					}
					sleepFor := time.Duration(sleepSecs) * time.Second
					for {
						if clearAt == nil {
							// timer was already reset?
							if debug {
								log.Printf("clearAfter: timer was reset, skipping\n")
							}
							clearLoopRunning = false
							return
						}

						if time.Now().After(*clearAt) {
							lock.Lock()
							defer lock.Unlock()
							if debug {
								log.Printf("clearAfter: has been %d seconds since clipboard was last set, clearing clipboard\n", clearAfter)
							}
							clipboard = Clipboard{Text: "", Timestamp: time.Now()}
							clearAt = nil
							clearLoopRunning = false
							return
						}
						if debug {
							log.Printf("clearAfter: clear loop sleeping for %d seconds\n", sleepSecs)
						}
						time.Sleep(sleepFor)
					}
				}()
				clearLoopRunning = true
			}
		}
	}()

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			// write index to resposne
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
		if debug {
			log.Printf("copying '%s' to clipboard\n", input.Text)
		}
		clipboard = Clipboard{Text: input.Text, Timestamp: time.Now()}

		w.Header().Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated remote clipboard"))

		// clear clipboard after x seconds
		if clearAfter > 0 {
			clearer <- true
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
