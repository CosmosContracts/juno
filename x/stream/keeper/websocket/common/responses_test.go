package common_test

import (
	"testing"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/stretchr/testify/require"
)

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected map[string]string
	}{
		{
			name:     "simple error message",
			message:  "something went wrong",
			expected: map[string]string{"error": "something went wrong"},
		},
		{
			name:     "empty error message",
			message:  "",
			expected: map[string]string{"error": ""},
		},
		{
			name:     "error with special characters",
			message:  "error: invalid address \"test@123\"",
			expected: map[string]string{"error": "error: invalid address \"test@123\""},
		},
		{
			name:     "very long error message",
			message:  string(make([]byte, 1000, 1000)),
			expected: map[string]string{"error": string(make([]byte, 1000, 1000))},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := common.ErrorResponse(tc.message)
			require.Equal(t, tc.expected, response)
			require.Len(t, response, 1)
			require.Contains(t, response, "error")
		})
	}
}

func TestDataResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		validate func(t *testing.T, response map[string]any)
	}{
		{
			name: "string data",
			data: "test data",
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, "test data", response["data"])
			},
		},
		{
			name: "numeric data",
			data: 42,
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, 42, response["data"])
			},
		},
		{
			name: "nil data",
			data: nil,
			validate: func(t *testing.T, response map[string]any) {
				require.Nil(t, response["data"])
			},
		},
		{
			name: "struct data",
			data: struct {
				Name  string
				Value int
			}{Name: "test", Value: 100},
			validate: func(t *testing.T, response map[string]any) {
				data := response["data"].(struct {
					Name  string
					Value int
				})
				require.Equal(t, "test", data.Name)
				require.Equal(t, 100, data.Value)
			},
		},
		{
			name: "map data",
			data: map[string]int{"a": 1, "b": 2},
			validate: func(t *testing.T, response map[string]any) {
				data := response["data"].(map[string]int)
				require.Equal(t, 1, data["a"])
				require.Equal(t, 2, data["b"])
			},
		},
		{
			name: "slice data",
			data: []string{"one", "two", "three"},
			validate: func(t *testing.T, response map[string]any) {
				data := response["data"].([]string)
				require.Len(t, data, 3)
				require.Equal(t, "one", data[0])
				require.Equal(t, "two", data[1])
				require.Equal(t, "three", data[2])
			},
		},
		{
			name: "boolean data",
			data: true,
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["data"])
			},
		},
		{
			name: "empty string data",
			data: "",
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, "", response["data"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := common.DataResponse(tc.data)
			require.Len(t, response, 1)
			require.Contains(t, response, "data")
			tc.validate(t, response)
		})
	}
}

func TestBalancesResponse(t *testing.T) {
	tests := []struct {
		name     string
		balances any
		validate func(t *testing.T, response map[string]any)
	}{
		{
			name: "empty balances",
			balances: []any{},
			validate: func(t *testing.T, response map[string]any) {
				balances := response["balances"].([]any)
				require.Len(t, balances, 0)
			},
		},
		{
			name: "single balance",
			balances: map[string]string{"denom": "ujuno", "amount": "1000"},
			validate: func(t *testing.T, response map[string]any) {
				balance := response["balances"].(map[string]string)
				require.Equal(t, "ujuno", balance["denom"])
				require.Equal(t, "1000", balance["amount"])
			},
		},
		{
			name: "multiple balances",
			balances: []map[string]string{
				{"denom": "ujuno", "amount": "1000"},
				{"denom": "uatom", "amount": "500"},
			},
			validate: func(t *testing.T, response map[string]any) {
				balances := response["balances"].([]map[string]string)
				require.Len(t, balances, 2)
				require.Equal(t, "ujuno", balances[0]["denom"])
				require.Equal(t, "1000", balances[0]["amount"])
				require.Equal(t, "uatom", balances[1]["denom"])
				require.Equal(t, "500", balances[1]["amount"])
			},
		},
		{
			name: "nil balances",
			balances: nil,
			validate: func(t *testing.T, response map[string]any) {
				require.Nil(t, response["balances"])
			},
		},
		{
			name: "complex balance structure",
			balances: struct {
				Total     string
				Available string
				Locked    string
			}{
				Total:     "1000",
				Available: "800",
				Locked:    "200",
			},
			validate: func(t *testing.T, response map[string]any) {
				balance := response["balances"].(struct {
					Total     string
					Available string
					Locked    string
				})
				require.Equal(t, "1000", balance.Total)
				require.Equal(t, "800", balance.Available)
				require.Equal(t, "200", balance.Locked)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := common.BalancesResponse(tc.balances)
			require.Len(t, response, 1)
			require.Contains(t, response, "balances")
			tc.validate(t, response)
		})
	}
}

func TestNotFoundResponse(t *testing.T) {
	response := common.NotFoundResponse()
	
	require.Len(t, response, 1)
	require.Contains(t, response, "found")
	require.Equal(t, false, response["found"])
	
	// Test multiple calls return consistent result
	response2 := common.NotFoundResponse()
	require.Equal(t, response, response2)
}

func TestFoundResponse(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		data     any
		validate func(t *testing.T, response map[string]any)
	}{
		{
			name: "simple string data",
			key:  "result",
			data: "found it",
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				require.Equal(t, "found it", response["result"])
			},
		},
		{
			name: "delegation data",
			key:  "delegation",
			data: map[string]string{
				"delegator": "juno1...",
				"validator": "junovaloper1...",
				"shares":    "1000000",
			},
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				delegation := response["delegation"].(map[string]string)
				require.Equal(t, "juno1...", delegation["delegator"])
				require.Equal(t, "junovaloper1...", delegation["validator"])
				require.Equal(t, "1000000", delegation["shares"])
			},
		},
		{
			name: "empty key",
			key:  "",
			data: "data",
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				require.Equal(t, "data", response[""])
			},
		},
		{
			name: "nil data",
			key:  "empty",
			data: nil,
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				require.Nil(t, response["empty"])
			},
		},
		{
			name: "numeric data",
			key:  "count",
			data: 42,
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				require.Equal(t, 42, response["count"])
			},
		},
		{
			name: "array data",
			key:  "items",
			data: []int{1, 2, 3},
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				items := response["items"].([]int)
				require.Equal(t, []int{1, 2, 3}, items)
			},
		},
		{
			name: "special characters in key",
			key:  "special-key_123",
			data: "value",
			validate: func(t *testing.T, response map[string]any) {
				require.Equal(t, true, response["found"])
				require.Equal(t, "value", response["special-key_123"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := common.FoundResponse(tc.key, tc.data)
			require.Len(t, response, 2)
			require.Contains(t, response, "found")
			require.Contains(t, response, tc.key)
			tc.validate(t, response)
		})
	}
}

func TestResponseConsistency(t *testing.T) {
	// Test that responses are consistent for the same input
	t.Run("error response consistency", func(t *testing.T) {
		msg := "test error"
		resp1 := common.ErrorResponse(msg)
		resp2 := common.ErrorResponse(msg)
		require.Equal(t, resp1, resp2)
	})

	t.Run("data response consistency", func(t *testing.T) {
		data := map[string]int{"a": 1, "b": 2}
		resp1 := common.DataResponse(data)
		resp2 := common.DataResponse(data)
		require.Equal(t, resp1, resp2)
	})

	t.Run("not found response consistency", func(t *testing.T) {
		resp1 := common.NotFoundResponse()
		resp2 := common.NotFoundResponse()
		require.Equal(t, resp1, resp2)
	})
}

func TestResponseImmutability(t *testing.T) {
	// Test that modifying returned response doesn't affect future calls
	t.Run("error response immutability", func(t *testing.T) {
		resp1 := common.ErrorResponse("test")
		resp1["error"] = "modified"
		resp1["extra"] = "field"
		
		resp2 := common.ErrorResponse("test")
		require.Equal(t, "test", resp2["error"])
		require.NotContains(t, resp2, "extra")
	})

	t.Run("not found response immutability", func(t *testing.T) {
		resp1 := common.NotFoundResponse()
		resp1["found"] = true
		resp1["extra"] = "field"
		
		resp2 := common.NotFoundResponse()
		require.Equal(t, false, resp2["found"])
		require.NotContains(t, resp2, "extra")
	})
}