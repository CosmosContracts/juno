package app_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/CosmosContracts/juno/v31/testutil/setup"
)

func TestLegacyParameterChangeProposalCanBeUnpackedForExport(t *testing.T) {
	app := setup.Setup(false, t.TempDir(), "juno-1", t)
	legacyProposal := &paramsproposal.ParameterChangeProposal{
		Title:       "legacy parameter change",
		Description: "state retained from before x/params removal",
		Changes: []paramsproposal.ParamChange{{
			Subspace: "staking",
			Key:      "MaxValidators",
			Value:    "100",
		}},
	}
	anyProposal, err := codectypes.NewAnyWithValue(legacyProposal)
	require.NoError(t, err)
	encoded := app.AppCodec().MustMarshal(anyProposal)
	var storedProposal codectypes.Any
	app.AppCodec().MustUnmarshal(encoded, &storedProposal)

	var content govv1beta1.Content
	require.NoError(t, app.InterfaceRegistry().UnpackAny(&storedProposal, &content))
	require.Equal(t, legacyProposal, content)
}

func TestLegacyGlobalFeeMessageCanBeUnpackedForExport(t *testing.T) {
	app := setup.Setup(false, t.TempDir(), "juno-1", t)
	storedMessage := &codectypes.Any{
		TypeUrl: "/gaia.globalfee.v1beta1.MsgUpdateParams",
		Value:   []byte{},
	}

	var message sdk.Msg
	require.NoError(t, app.InterfaceRegistry().UnpackAny(storedMessage, &message))
}

func TestLegacyBuilderMessageCanBeUnpackedForExport(t *testing.T) {
	app := setup.Setup(false, t.TempDir(), "juno-1", t)
	storedMessage := &codectypes.Any{
		TypeUrl: "/pob.builder.v1.MsgUpdateParams",
		Value:   []byte{},
	}

	var message sdk.Msg
	require.NoError(t, app.InterfaceRegistry().UnpackAny(storedMessage, &message))
}
