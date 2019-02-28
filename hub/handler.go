package hub

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/lovoo/goka"
)

func SetVisibility(visible bool) Handler {
	return func(ctx goka.Context, m *shared.HubContext) error {
		log.Infof("set visibility to %T:%+v", visible, m)

		exec := shared.NewExecutor(ctx)
		return exec.ApplyContext(m.ReceiverID, shared.Properties{
			"visible": visible,
		})
	}
}

func IsNodeInvisible(arg interface{}) bool {
	ctx := arg.(shared.HubContext)
	return ctx.Properties.MustBool("visible") == false
}
