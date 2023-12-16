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

// TODO:
// https://github.com/rogchap/v8go ?
// Allow ZIP file upload into the contract? (easier to use images, else you need to use IPFS or direct links off chain)

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
			if _, err := w.Write([]byte("contract and name are required")); err != nil {
				fmt.Println("error writing response: ", err)
			}
			return
		}

		contract := params["contract"][0]
		name := params["name"][0]

		b64Query := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"get_website":{"name":"%s"}}`, name)))

		res, err := http.Get(restServer + contract + "/smart/" + b64Query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeResponse(w, "error creating request: "+err.Error())
			return
		}
		defer res.Body.Close()

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
			os.Exit(1)
		}

		if len(resBody) > 0 && strings.Contains(string(resBody), "data") {
			var website Website
			err = json.Unmarshal(resBody, &website)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				writeResponse(w, "error unmarshalling response body: "+err.Error())
				return
			}

			writeResponse(w, website.Data.Source)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		writeResponse(w, "no website found")
	}
}

func writeResponse(w http.ResponseWriter, resBody string) {
	if _, err := w.Write([]byte(resBody)); err != nil {
		fmt.Println("error writing response: ", err)
	}
}
