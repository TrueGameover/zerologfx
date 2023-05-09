package types

import (
	"context"
	"github.com/TrueGameover/zerologfx/src/public"
)

type ZeroLogFxModule struct {
	Config public.ModuleConfig
	AppCtx context.Context
}
