package hub

import (
	"context"

	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/kafka"
)

func handleHubEvents(ctx goka.Context, msg interface{}, actions ...Action) error {
	m, ok := msg.(*shared.HubContext)
	if !ok {
		return errors.Errorf("invalid message type %+v", msg)
	}

	for _, action := range actions {
		handled, err := action.applyMessage(ctx, m)
		if err != nil {
			return errors.Annotate(err, "applyMessage")
		}

		if !handled {
			log.Warningf("unhandled hub msg [no match]: %+v", m)
		}
	}

	return nil
}

// receives dedicated hub messages, sends hub messages
func CreateConsumerDefaults(descr shared.EntityDescriptor, actions ...Action) shared.DispatcherFunc {
	act := []Action{}
	for _, action := range actions {
		act = append(act, action.setDescriptor(descr))
	}
	return CreateConsumer(
		descr.HubGroup(),
		descr.HubInputStream(),
		descr.HubOutputStream(),
		act...,
	)
}

func CreateConsumer(group goka.Group, inputStream, outputStream goka.Stream, actions ...Action) shared.DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(shared.HubContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleHubEvents(ctx, msg, actions...); err != nil {
						log.Error(errors.Annotate(err, "handleHubEvents"))
					}
				}),
				goka.Output(outputStream, new(shared.HubContextCodec)),
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
