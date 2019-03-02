package event

import (
	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
)

func NotifySuperOrdinates() Handler {
	return func(ctx goka.Context, descr shared.EntityDescriptor, m *shared.EventContext) error {
		log.Infof("notify superordinates:%v", m)

		exec := shared.NewExecutor(ctx)
		if err := exec.NotifySuperOrdinates(descr.Label(), m.NodeID,
			shared.UpdatedOperation, m.Properties,
		); err != nil {
			return errors.Annotate(err, "NotifySuperOrdinates")
		}

		return nil
	}
}
