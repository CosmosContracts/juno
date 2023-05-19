package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

// Move this into ictest under 'cosmos' package.

type GenesisKV struct {
	Key   string      `json:"key"`
	Value interface{} `json:"val"`
}

func ModifyGenesis(genesisKV []GenesisKV) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}

		for idx, values := range genesisKV {
			path := strings.Split(values.Key, ".")

			result := make([]interface{}, len(path))
			for i, component := range path {
				if v, err := strconv.Atoi(component); err == nil {
					result[i] = v
				} else {
					result[i] = component
				}
			}

			if err := dyno.Set(g, values.Value, result...); err != nil {
				return nil, fmt.Errorf("failed to set value (index:%d) in genesis json: %w", idx, err)
			}
		}

		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}
