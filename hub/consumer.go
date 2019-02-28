package hub

import (
	"context"
	"fmt"

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
		handled, err := action.ApplyMessage(ctx, m)
		if err != nil {
			return errors.Annotate(err, "ApplyMessage")
		}

		if !handled {
			log.Warningf("unhandled hub msg [no match]: %+v", m)
		}
	}

	return nil
}

func CreateConsumerDefaults(label string, actions ...Action) shared.DispatcherFunc {
	group := goka.Group(fmt.Sprintf("%s_HubGroup", label))
	outputStream := goka.Stream(fmt.Sprintf("Hub2%s", label))
	return CreateConsumer(group, shared.HubStream, outputStream, actions...)
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
