package webhost

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	address    = ":8787"
	restServer = "http://0.0.0.0:1317/cosmwasm/wasm/v1/contract/"
)

type Website struct {
	Data struct {
		Creator string `json:"creator"`
		Source  string `json:"source"`
	} `json:"data"`
}

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get request route
		route := strings.Split(r.URL.Path, "/")[2:]

		// Pull out contract & name from route
		contract := route[0]
		name := route[1]

		// Determine file path, default to index.html if none specified
		path := strings.Join(route[2:], "/")
		if path == "" {
			path = "index.html"
		}

		// Send request to CosmWasm contract, retrieve website
		website, err := makeRequest(contract, name, path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeResponse(w, "error making request: "+err.Error())
			return
		}

		// Decode & unzip website source files
		zr, err := decodeAndUnzip(website.Data.Source)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeResponse(w, "error decoding and unzipping file: "+err.Error())
			return
		}

		// Find file in zip at path
		for _, f := range zr.File {
			if f.Name == path {
				rc, err := f.Open()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeResponse(w, "error opening file: "+err.Error())
					return
				}
				defer rc.Close()

				// Read file into buffer
				buf := new(bytes.Buffer)
				if _, err := io.Copy(buf, rc); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeResponse(w, "error reading file: "+err.Error())
					return
				}

				// Set content type
				setContentType(w, path)

				// Write buffer to response
				if _, err := w.Write(buf.Bytes()); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeResponse(w, "error writing response: "+err.Error())
				}

				break
			}
		}
	}
}

// Make request to contract through CosmWasm REST server
func makeRequest(contract, name, path string) (Website, error) {
	var website Website

	b64Query := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"get_website":{"name":"%s"}}`, name)))
	url := restServer + contract + "/smart/" + b64Query

	res, err := http.Get(url)
	if err != nil {
		return website, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return website, err
	}

	if len(resBody) > 0 && strings.Contains(string(resBody), "data") {
		if err = json.Unmarshal(resBody, &website); err != nil {
			return website, err
		}
	}

	return website, nil
}

// Decode from base 64 and unzip file, return zip reader
func decodeAndUnzip(source string) (*zip.Reader, error) {
	// Decode from base 64
	decodedBytes, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return nil, err
	}

	// Unzip
	reader := bytes.NewReader(decodedBytes)
	zr, err := zip.NewReader(reader, int64(len(decodedBytes)))
	if err != nil {
		return nil, err
	}

	return zr, nil
}

// Set the content type of the response based on the file's extension
func setContentType(w http.ResponseWriter, filePath string) {
	// Set content type
	if strings.HasSuffix(filePath, ".html") {
		w.Header().Set("Content-Type", "text/html")
	} else if strings.HasSuffix(filePath, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(filePath, ".js") {
		w.Header().Set("Content-Type", "text/javascript")
	} else if strings.HasSuffix(filePath, ".png") {
		w.Header().Set("Content-Type", "image/png")
	} else if strings.HasSuffix(filePath, ".jpg") || strings.HasSuffix(filePath, ".jpeg") {
		w.Header().Set("Content-Type", "image/jpeg")
	} else if strings.HasSuffix(filePath, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
	} else if strings.HasSuffix(filePath, ".svg") {
		w.Header().Set("Content-Type", "image/svg+xml")
	} else if strings.HasSuffix(filePath, ".ico") {
		w.Header().Set("Content-Type", "image/x-icon")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
}

// Helper function for writing response
func writeResponse(w http.ResponseWriter, resBody string) {
	if _, err := w.Write([]byte(resBody)); err != nil {
		fmt.Println("error writing response: ", err)
	}
}
