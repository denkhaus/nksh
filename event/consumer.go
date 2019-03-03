package event

import (
	"context"

	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/kafka"
)

func handleInputEvents(ctx goka.Context, msg interface{}, exes ...Executable) error {
	m, ok := msg.(*shared.EventContext)
	if !ok {
		return errors.Errorf("invalid message type %+v", msg)
	}

	for _, exe := range exes {
		if state := exe.Execute(ctx, m); state.Failed() {
			log.Warningf("unhandled input msg [%s]: %+v", state, m)
		}
	}

	return nil
}

// receives input messages, sends hub messages
func CreateConsumerDefaults(descr shared.EntityDescriptor, execs ...Executable) shared.DispatcherFunc {
	exe := []Executable{}
	for _, exec := range execs {
		exe = append(exe, exec.SetDescriptor(descr))
	}
	return CreateConsumer(
		descr.EventGroup(),
		descr.EventInputStream(),
		descr.EventOutputStream(),
		exe...,
	)
}

func CreateConsumer(group goka.Group, inputStream, outputStream goka.Stream, execs ...Executable) shared.DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(shared.EventContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleInputEvents(ctx, msg, execs...); err != nil {
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
