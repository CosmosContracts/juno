package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"reflect"
	"strings"
)

type MainConfig struct {
	Chains []Chain `json:"chains"`
}

type Chain struct {
	// ibc chain config (optional)
	ChainType      string `json:"chain-type"`
	CoinType       int    `json:"coin-type"`
	Binary         string `json:"binary"`
	Bech32Prefix   string `json:"bech32-prefix"`
	Denom          string `json:"denom"`
	TrustingPeriod string `json:"trusting-period"`
	Debugging      bool   `json:"debugging"`

	// Required
	Name    string `json:"name"`
	ChainID string `json:"chain-id"`

	DockerImage struct {
		Repository string `json:"repository"`
		Version    string `json:"version"`
		UidGid     string `json:"uid-gid"`
	} `json:"docker-image"`

	GasPrices     string  `json:"gas-prices"`
	GasAdjustment float64 `json:"gas-adjustment"`
	NumberVals    int     `json:"number-vals"`
	NumberNode    int     `json:"number-node"`
	BlocksTTL     int     `json:"blocks-ttl"`
	IBCPath       string  `json:"ibc-path"`
	Genesis       Genesis `json:"genesis"`
}

type Genesis struct {
	Modify []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"modify"`
	Accounts []struct {
		Name     string `json:"name"`
		Amount   string `json:"amount"`
		Address  string `json:"address"`
		Mnemonic string `json:"mnemonic"`
	} `json:"accounts"`
}

type LogOutput struct {
	ChainID     string `json:"chain-id"`
	ChainName   string `json:"chain-name"`
	RPCAddress  string `json:"rpc-address"`
	GRPCAddress string `json:"grpc-address"`
	IBCPath     string `json:"ibc-path"`
}

func LoadConfig() (*MainConfig, error) {
	// read from current dir "config.json". Allow user ability for multiple configs
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var config MainConfig
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	// Defaults
	for i := range config.Chains {
		chain := &config.Chains[i]

		if chain.BlocksTTL <= 0 {
			chain.BlocksTTL = math.MaxInt32
		}

		if chain.ChainType == "" {
			chain.ChainType = "cosmos"
		}

		if chain.CoinType == 0 {
			chain.CoinType = 118
		}

		if chain.DockerImage.UidGid == "" {
			chain.DockerImage.UidGid = "1025:1025"
		}

		// TODO: Error here instead?
		if chain.Binary == "" {
			chain.Binary = "junod"
		}
		if chain.Denom == "" {
			chain.Denom = "ujuno"
		}
		if chain.Bech32Prefix == "" {
			chain.Bech32Prefix = "juno"
		}

		if chain.TrustingPeriod == "" {
			chain.TrustingPeriod = "112h"
		}
	}

	// Replace env variables
	for i := range config.Chains {
		chain := config.Chains[i]
		replaceStringValues(&chain, "%DENOM%", chain.Denom)

		config.Chains[i] = chain
	}

	return &config, nil
}

func replaceStringValues(data interface{}, oldStr, replacement string) {
	replaceStringFields(reflect.ValueOf(data), oldStr, replacement)
}

func replaceStringFields(value reflect.Value, oldStr, replacement string) {
	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return
		}
		replaceStringFields(value.Elem(), oldStr, replacement)
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			replaceStringFields(field, oldStr, replacement)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			replaceStringFields(value.Index(i), oldStr, replacement)
		}
	case reflect.String:
		currentStr := value.String()
		if strings.Contains(currentStr, oldStr) {
			updatedStr := strings.Replace(currentStr, oldStr, replacement, -1)
			value.SetString(updatedStr)
		}
	}
}
