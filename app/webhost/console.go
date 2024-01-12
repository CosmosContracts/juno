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
	address             = ":8787"
	restServer          = "http://0.0.0.0:1317/cosmwasm/wasm/v1/contract/"
	maxZipFileSize      = uint64(1024 * 1024 * 5) // 5 MB
	maxUnzippedFileSize = uint64(1024 * 1024 * 1) // 1 MB
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
		website, err := makeRequest(contract, name)
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

				// Check if file size exceeds maximum allowed size
				if f.UncompressedSize64 > maxUnzippedFileSize {
					w.WriteHeader(http.StatusInternalServerError)
					writeResponse(w, "resource exceeds maximum allowed size of "+fmt.Sprintf("%d", maxUnzippedFileSize)+" bytes")
					return
				}

				// Define limit to prevent linting error (we check for file size limit above)
				maxSize := int64(f.UncompressedSize64)

				// Copy file to buffer
				if _, err := io.CopyN(buf, rc, maxSize); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					writeResponse(w, "error reading file: "+err.Error())
					return
				}

				// Set content type based on file type
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
func makeRequest(contract string, name string) (Website, error) {
	var website Website

	b64Query := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"get_website":{"name":"%s"}}`, name)))

	res, err := http.Get(restServer + contract + "/smart/" + b64Query)
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

	// Check if decoded content exceeds the maximum allowed size
	if uint64(len(decodedBytes)) > maxZipFileSize {
		return nil, fmt.Errorf("decoded content exceeds maximum allowed size of %d bytes", maxZipFileSize)
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
	switch {
	case strings.HasSuffix(filePath, ".html"):
		w.Header().Set("Content-Type", "text/html")
	case strings.HasSuffix(filePath, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(filePath, ".js"):
		w.Header().Set("Content-Type", "text/javascript")
	case strings.HasSuffix(filePath, ".png"):
		w.Header().Set("Content-Type", "image/png")
	case strings.HasSuffix(filePath, ".jpg"), strings.HasSuffix(filePath, ".jpeg"):
		w.Header().Set("Content-Type", "image/jpeg")
	case strings.HasSuffix(filePath, ".gif"):
		w.Header().Set("Content-Type", "image/gif")
	case strings.HasSuffix(filePath, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	case strings.HasSuffix(filePath, ".ico"):
		w.Header().Set("Content-Type", "image/x-icon")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
}

// Helper function for writing response
func writeResponse(w http.ResponseWriter, resBody string) {
	if _, err := w.Write([]byte(resBody)); err != nil {
		fmt.Println("error writing response: ", err)
	}
}
