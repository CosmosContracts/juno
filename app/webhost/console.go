package webhost

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// https://github.com/rogchap/v8go ?

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
		params := r.URL.Query()

		if len(params["contract"]) == 0 || len(params["name"]) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("contract and name are required"))
			return
		}

		contract := params["contract"][0]
		name := params["name"][0]

		b64Query := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"get_website":{"name":"%s"}}`, name)))
		// fmt.Println("b64Query is: ", b64Query)

		res, err := http.Get(restServer + contract + "/smart/" + b64Query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error creating request: " + err.Error()))
			return
		}
		defer res.Body.Close()

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}
		// fmt.Printf("client: response body: %s\n", resBody)

		if len(resBody) > 0 && strings.Contains(string(resBody), "data") {
			var website Website
			err = json.Unmarshal(resBody, &website)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error unmarshalling response body: " + err.Error()))
				return
			}

			fmt.Println("website is: ", website)

			w.Write([]byte(website.Data.Source))
			return
		}

		// error, no website body found
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no website found"))
	}
}

// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		params := r.URL.Query()

// 		if len(params["contract"]) == 0 || len(params["name"]) == 0 {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte("contract and name are required"))
// 			return
// 		}

// 		contract := params["contract"][0]
// 		name := params["name"][0]

// 		b64Query := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"get_website":{"name":"%s"}}`, name)))
// 		// fmt.Println("b64Query is: ", b64Query)

// 		res, err := http.Get(restServer + contract + "/smart/" + b64Query)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte("error creating request: " + err.Error()))
// 			return
// 		}
// 		defer res.Body.Close()

// 		resBody, err := io.ReadAll(res.Body)
// 		if err != nil {
// 			fmt.Printf("client: could not read response body: %s\n", err)
// 			os.Exit(1)
// 		}
// 		// fmt.Printf("client: response body: %s\n", resBody)

// 		if len(resBody) > 0 && strings.Contains(string(resBody), "data") {
// 			var website Website
// 			err = json.Unmarshal(resBody, &website)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte("error unmarshalling response body: " + err.Error()))
// 				return
// 			}

// 			fmt.Println("website is: ", website)

// 			w.Write([]byte(website.Data.Source))
// 			return
// 		}

// 		// error, no website body found
// 		w.WriteHeader(http.StatusNotFound)
// 		w.Write([]byte("no website found"))
// 	})

// 	http.ListenAndServe(address, nil)

// }

// func servezip(res http.ResponseWriter, req *http.Request) {
// 	zippath := "files/" + strings.Split(html.EscapeString(req.URL.Path), "/")[2] + ".zip"

// 	z, err := zip.OpenReader(zippath)
// 	if err != nil {
// 		http.Error(res, err.Error(), 404)
// 		return
// 	}
// 	defer z.Close()
// 	http.StripPrefix("/zip/", http.FileServer(zipfs.NewZipFS(&z.Reader)))
// }

// http.StripPrefix("/zip/", http.FileServer(zipfs.NewZipFS(&z.Reader))).ServeHTTP(res, req)
