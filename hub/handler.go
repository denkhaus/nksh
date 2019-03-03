package hub

import (
	"github.com/denkhaus/nksh/shared"
)

func SetVisibility(visible bool) shared.Handler {
	return func(ctx *shared.HandlerContext) error {
		log.Infof("set visibility to %t:%+v", visible, ctx.HubContext)

		exec := shared.NewExecutor(ctx)
		return exec.ApplyProperties(ctx.HubContext.ReceiverID,
			shared.Properties{
				"visible": visible,
			})
	}
}

func IsNodeInvisible(arg interface{}) bool {
	ctx := arg.(shared.HubContext)
	return ctx.Properties.MustBool("visible") == false
}
