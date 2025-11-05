package encoding

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestBuildRequestMessageHandlesCamelCaseKeys(t *testing.T) {
	desc := &MethodDescriptor{RequestType: "cosmos.staking.v1beta1.QueryDelegatorValidatorsRequest"}
	delegator := "juno1delegatorxyz"
	msg, err := BuildRequestMessage(desc, map[string]string{
		"delegatorAddr": delegator,
	})
	require.NoError(t, err)
	req, ok := msg.(*stakingtypes.QueryDelegatorValidatorsRequest)
	require.True(t, ok)
	require.Equal(t, delegator, req.DelegatorAddr)
}

func TestBuildRequestMessageHandlesDottedNestedKeys(t *testing.T) {
	desc := &MethodDescriptor{RequestType: "cosmos.bank.v1beta1.QueryDenomOwnersRequest"}
	key := base64.StdEncoding.EncodeToString([]byte("next"))
	msg, err := BuildRequestMessage(desc, map[string]string{
		"denom":            "ujuno",
		"pagination.key":   key,
		"pagination.limit": "25",
	})
	require.NoError(t, err)
	req, ok := msg.(*banktypes.QueryDenomOwnersRequest)
	require.True(t, ok)
	require.Equal(t, "ujuno", req.Denom)
	require.NotNil(t, req.Pagination)
	require.Equal(t, []byte("next"), req.Pagination.Key)
	require.EqualValues(t, 25, req.Pagination.Limit)
}
