package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
)

func NotifySuperOrdinates() shared.Handler {
	return func(ctx *shared.HandlerContext) error {
		log.Infof("notify superordinates: %v", ctx.EventContext)

		exec := shared.NewExecutor(ctx)
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

func LoadEntityContext() shared.Handler {
	return func(ctx *shared.HandlerContext) (err error) {
		exec := shared.NewExecutor(ctx)
		entityCtx, err := exec.BuildEntityContext(
			ctx.EventContext.NodeID,
		)
		if err != nil {
			return errors.Annotate(err, "BuildEntityContext")
		}

		ctx.Set("EntityContext", entityCtx)
		return nil
	}
}
