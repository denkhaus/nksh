package event

import (
	"context"

	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/kafka"
)

func handleInputEvents(ctx goka.Context, msg interface{}, actions ...Action) error {
	m, ok := msg.(*shared.EventContext)
	if !ok {
		return errors.Errorf("invalid message type %+v", msg)
	}

	for _, action := range actions {
		handled, err := action.applyContext(ctx, m)
		if err != nil {
			return errors.Annotate(err, "ApplyMessage")
		}

		if !handled {
			log.Warningf("unhandled input msg [no match]: %+v", m)
		}
	}

	return nil
}

// receives input messages, sends hub messages
func CreateConsumerDefaults(descr shared.EntityDescriptor, actions ...Action) shared.DispatcherFunc {
	act := []Action{}
	for _, action := range actions {
		act = append(act, action.setDescriptor(descr))
	}
	return CreateConsumer(
		descr.EventGroup(),
		descr.EventInputStream(),
		descr.EventOutputStream(),
		act...,
	)
}

func CreateConsumer(group goka.Group, inputStream, outputStream goka.Stream, actions ...Action) shared.DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(shared.EventContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleInputEvents(ctx, msg, actions...); err != nil {
						log.Error(errors.Annotate(err, "handleInputEvents"))
					}
				}), goka.Output(outputStream, new(shared.EventContextCodec)),
			)

			p, err := goka.NewProcessor(kServers, g,
				goka.WithTopicManagerBuilder(
					kafka.ZKTopicManagerBuilder(zServers),
				),
			)
			if err != nil {
				return errors.Annotate(err, "NewProcessor")
			}

			if err := p.Run(ctx); err != nil {
				return errors.Annotate(err, "Run")
			}

			return nil
		}
	}
}
