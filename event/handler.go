package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
)

func NotifySuperOrdinates() Handler {
	return func(ctx *HandlerContext) error {
		log.Infof("notify superordinates: %v", ctx.EventContext)

		exec := shared.NewExecutor(ctx.GokaContext)
		if err := exec.NotifySuperOrdinates(
			ctx.EntityDescriptor.Label(),
			ctx.EventContext.NodeID,
			shared.UpdatedOperation,
			ctx.EventContext.Properties,
		); err != nil {
			return errors.Annotate(err, "NotifySuperOrdinates")
		}

		return nil
	}
}
