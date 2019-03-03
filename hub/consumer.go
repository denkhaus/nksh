package hub

import (
	"context"

	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/kafka"
)

func handleHubEvents(ctx goka.Context, msg interface{}, exes ...Executable) error {
	m, ok := msg.(*shared.HubContext)
	if !ok {
		return errors.Errorf("invalid message type %+v", msg)
	}

	for _, exe := range exes {
		if state := exe.Execute(ctx, m); state.Failed() {
			log.Warningf("unhandled hub msg [%s]: %+v", state, m)
		}
	}

	return nil
}

// receives dedicated hub messages, sends hub messages
func CreateConsumerDefaults(descr shared.EntityDescriptor, execs ...Executable) shared.DispatcherFunc {
	exe := []Executable{}
	for _, exec := range execs {
		exe = append(exe, exec.SetDescriptor(descr))
	}
	return CreateConsumer(
		descr.HubGroup(),
		descr.HubInputStream(),
		descr.HubOutputStream(),
		exe...,
	)
}

func CreateConsumer(group goka.Group, inputStream, outputStream goka.Stream, execs ...Executable) shared.DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(shared.HubContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleHubEvents(ctx, msg, execs...); err != nil {
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
