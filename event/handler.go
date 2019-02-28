package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
)

func NotifySuperOrdinates(entityLabel string) Handler {
	return func(ctx goka.Context, m *shared.EventContext) error {
		log.Infof("notify superordinates:%+v", m)

		exec := shared.NewExecutor(ctx)
		if err := exec.NotifySuperOrdinates(entityLabel, m.NodeID,
			shared.UpdatedOperation, m.Properties,
		); err != nil {
			return errors.Annotate(err, "NotifySuperOrdinates")
		}

		return nil
	}
}
