package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// verify interface at compile time
var (
	_ sdk.Msg = &MsgUpdateParams{}
)
