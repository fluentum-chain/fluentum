package main

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

var _ servertypes.AppCreator = (*appCreator)(nil)
var _ servertypes.AppExporter = (*appCreator)(nil)
