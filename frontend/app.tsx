import React from "react";
import ReactDOM from "react-dom";

import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";

function copyToClipboard(text: string): Promise<void> {
  return navigator.clipboard.writeText(text);
}

function urlJoin(baseURL: string, path: string): string {
  if (!baseURL.endsWith("/")) {
    baseURL += "/";
  }
  return baseURL + path;
}

const App = () => {
  const [password, setPassword] = React.useState("");
  const [serverUrl, setServerUrl] = React.useState("");
  const [copyText, setCopyText] = React.useState("");

  const defaultServerUrl = window.location.href;

  const copy = (text: string) => {
    // default to defaultServerUrl if serverUrl is empty
    const url = serverUrl.trim() || defaultServerUrl;
    fetch(urlJoin(url, "copy"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        password: password,
      },
      body: JSON.stringify({ text }),
    }).then((response) => {
      if (response.status === 200) {
        toast.success("Copied to server clipboard");
      } else {
        toast.error(
          `Failed to copy to server clipboard, status code: ${response.status} ${response.statusText}`
        );
      }
    });
  };

  const paste = async () => {
    // default to defaultServerUrl if serverUrl is empty
    const url = serverUrl.trim() || defaultServerUrl;
    fetch(urlJoin(url, "paste"), {
      method: "GET",
      headers: {
        password: password,
      },
    })
      .then((response) => {
        if (response.status === 200) {
          return response.text();
        } else {
          toast.error(
            `Failed to paste from server clipboard, status code: ${response.status} ${response.statusText}`
          );
        }
      })
      .then((text) => {
        if (text) {
          copyToClipboard(text)
            .then(() => {
              toast.success("Pasted from server into your clipboard");
            })
            .catch((err) => {
              toast.error(`Failed to paste into your clipboard: ${err}`);
              alert(text);
            });
        }
      });
  };

  return (
    <>
      <head>
        <title>Server Clipboard Frontend</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="charset" content="utf-8" />
        <link
          rel="stylesheet"
          href="https://cdn.simplecss.org/simple.min.css"
        />
        <style>
          {`
            button {
              padding: 1rem;
              font-size: 1.5rem;
            }

            label {
              display: flex;
              flex-direction: column;
              margin-bottom: 0.5rem;
              flex-direction: column;
            }
            caption {
              text-align: left;
              font-size: 0.8rem;
            }
          `}
        </style>
      </head>
      <body>
        <div>
          <h2>Server Clipboard</h2>
          <div id="configuration">
            <h3>Configuration</h3>
            <form onSubmit={() => {}}>
              <label>
                <span>Server URL:</span>
                <input
                  type="text"
                  id="server-url"
                  placeholder={defaultServerUrl}
                  value={serverUrl}
                  onChange={(e) => setServerUrl(e.target.value)}
                />
                <caption>Leave empty to use placeholder</caption>
              </label>
              <label>
                <span>Password:</span>
                <input
                  type="password"
                  id="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </label>
            </form>
          </div>
          <div>
            <div>
              <h3>Copy</h3>
              <textarea
                id="copy-text"
                placeholder="Text to copy"
                onChange={(e) => setCopyText(e.target.value)}
              ></textarea>

              <p>Click to copy text to the server</p>
              <button id="copy-button" onClick={() => copy(copyText)}>
                Copy to Server
              </button>
            </div>
            <div>
              <h3>Paste</h3>
              <p>Click to paste from server clipboard into yours</p>
              <button id="paste-button" onClick={paste}>
                Paste from Server
              </button>
            </div>
          </div>
        </div>
        <footer>
          <p>
            More Info at{" "}
            <a href="https://github.com/seanbreckenridge/server_clipboard">
              github.com/seanbreckenridge/server_clipboard
            </a>
          </p>
        </footer>
        <ToastContainer />
      </body>
    </>
  );
};
ReactDOM.render(<App />, document.getElementById("root"));
