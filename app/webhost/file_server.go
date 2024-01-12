package webhost

import (
	"archive/zip"
	"bytes"
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

		// Pull out path to file on contract
		path := strings.Join(route[1:], "/")
		if path == "" {
			path = "index.html"
		}
		fmt.Println("path: ", path)

		file := "/home/joel/development/web.zip"

		if file == "" {
			w.WriteHeader(http.StatusNotFound)
			writeResponse(w, "no website found")
			return
		}

		zr, err := zip.OpenReader(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeResponse(w, "error opening zip file: "+err.Error())
			return
		}

		defer zr.Close()

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
				w.WriteHeader(http.StatusOK)
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
