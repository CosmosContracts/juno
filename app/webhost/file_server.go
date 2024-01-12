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

func Handler2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get route
		route := strings.Split(r.URL.Path, "/")[2:]

		// Pull out contract
		contract := route[0]
		fmt.Println("contract: ", contract)

		// Pull out name
		name := route[1]
		fmt.Println("name: ", name)

		// Pull out path to file on contract
		path := strings.Join(route[2:], "/")
		if path == "" {
			path = "index.html"
		}
		fmt.Println("path: ", path)

		website, err := makeRequest(contract, name, path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeResponse(w, "error making request: "+err.Error())
			return
		}

		// Unzip website.Data.Source
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
					fmt.Println("error writing response: ", err)
				}

				break
			}
		}
	}
}

// Make request to contract
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

// Decode from base 64 and unzip file
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

// Function to set the header
func setContentType(w http.ResponseWriter, path string) {
	// Set content type
	if strings.HasSuffix(path, ".html") {
		w.Header().Set("Content-Type", "text/html")
	} else if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "text/javascript")
	} else if strings.HasSuffix(path, ".png") {
		w.Header().Set("Content-Type", "image/png")
	} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		w.Header().Set("Content-Type", "image/jpeg")
	} else if strings.HasSuffix(path, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
	} else if strings.HasSuffix(path, ".svg") {
		w.Header().Set("Content-Type", "image/svg+xml")
	} else if strings.HasSuffix(path, ".ico") {
		w.Header().Set("Content-Type", "image/x-icon")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
}
