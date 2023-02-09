package v13

import (
	"fmt"
	"strings"

	"github.com/CosmosContracts/juno/v13/app/keepers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	// oracletypes "github.com/CosmosContracts/juno/v13/x/oracle/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// Returns "ujunox" if the chain is uni, else returns the standard ujuno token denom.
func GetChainsDenomToken(chainID string) string {
	if strings.HasPrefix(chainID, "uni-") {
		return "ujunox"
	}
	return "ujuno"
}

func CreateV13_2UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// transfer module consensus version has been bumped to 2
		// the above is https://github.com/cosmos/ibc-go/blob/v5.1.0/docs/migrations/v3-to-v4.md
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		// TODO: must reAdd the keeper in app.go for this to work
		// Oracle
		// newOracleParams := oracletypes.DefaultParams()

		// // add osmosis to the oracle params
		// osmosisDenom := oracletypes.Denom{
		// 	BaseDenom:   "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
		// 	SymbolDenom: "OSMO",
		// 	Exponent:    uint32(6),
		// }

		// allDenoms := oracletypes.DefaultWhitelist
		// allDenoms = append(allDenoms, osmosisDenom)

		// newOracleParams.Whitelist = allDenoms
		// newOracleParams.TwapTrackingList = allDenoms
		// logger.Info(fmt.Sprintf("Oracle params set: %s", newOracleParams.String()))
		// keepers.OracleKeeper.SetParams(ctx, newOracleParams)

		return versionMap, err
	}
}
