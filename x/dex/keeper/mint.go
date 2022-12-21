// x/dex/keeper/mint.go

package keeper

import (
	"fmt"
	"interchange/x/dex/types"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
)

// isIBCToken checks if the token came from the IBC module
// Each IBC token starts with an ibc/ denom, the check is rather simple
func isIBCToken(denom string) bool {
	return strings.HasPrefix(denom, "ibc/")
}

func (k Keeper) SafeBurn(ctx sdk.Context, port string, channel string, sender sdk.AccAddress, denom string, amount int32) error {
	if isIBCToken(denom) {
		// Burn the tokens
		if err := k.BurnTokens(ctx, sender, sdk.NewCoin(denom, sdkmath.NewInt(int64(amount)))); err != nil {
			return err
		}
	} else {
		// Lock the tokens
		if err := k.LockTokens(ctx, port, channel, sender, sdk.NewCoin(denom, sdkmath.NewInt(int64(amount)))); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) BurnTokens(ctx sdk.Context, sender sdk.AccAddress, tokens sdk.Coin) error {
	// transfer the coins to the module account and burn them
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(tokens)); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(
		ctx, types.ModuleName, sdk.NewCoins(tokens),
	); err != nil {
		// NOTE: should not happen as the module account was
		// retrieved on the step above and it has enough balance
		// to burn.
		panic(fmt.Sprintf("cannot burn coins after a successful send to a module account: %v", err))
	}

	return nil
}

func (k Keeper) LockTokens(ctx sdk.Context, sourcePort string, sourceChannel string, sender sdk.AccAddress, tokens sdk.Coin) error {
	// create the escrow address for the tokens
	escrowAddress := ibctransfertypes.GetEscrowAddress(sourcePort, sourceChannel)

	// escrow source tokens. It fails if balance insufficient
	if err := k.bankKeeper.SendCoins(
		ctx, sender, escrowAddress, sdk.NewCoins(tokens),
	); err != nil {
		return err
	}

	return nil
}
