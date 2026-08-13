package ibc_test

import (
	"testing"

	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

func TestPFMEscrowAccountsUseEachHopsChannel(t *testing.T) {
	prefixes := [3]string{"juno-a", "juno-b", "juno-c"}
	abChan := &ibc.ChannelOutput{PortID: "transfer", ChannelID: "channel-0"}
	bcChan := ibc.ChannelCounterparty{PortID: "transfer", ChannelID: "channel-7"}
	cdChan := ibc.ChannelCounterparty{PortID: "transfer", ChannelID: "channel-42"}

	accounts := pfmEscrowAccounts(prefixes, abChan, bcChan, cdChan)

	channels := [3]ibc.ChannelCounterparty{
		{PortID: abChan.PortID, ChannelID: abChan.ChannelID},
		bcChan,
		cdChan,
	}
	for i, channel := range channels {
		expected := sdk.MustBech32ifyAddressBytes(
			prefixes[i],
			transfertypes.GetEscrowAddress(channel.PortID, channel.ChannelID),
		)
		require.Equal(t, expected, accounts[i], "hop %d must use its own channel ID", i+1)
	}
}
