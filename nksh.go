package nksh

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/kafka"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	log logrus.FieldLogger = logrus.New().WithField("package", "nksh")
)

type DispatcherFunc func(ctx context.Context, kServers, zServers []string) func() error

func Startup(kafkaHost, zookeeperHost string, funcs ...DispatcherFunc) error {
	kServers, err := LookupClusterHosts(kafkaHost, 9092)
	if err != nil {
		return errors.Annotate(err, "LookupClusterHosts [kafka]")
	}

	zServers, err := LookupClusterHosts(zookeeperHost, 2181)
	if err != nil {
		return errors.Annotate(err, "LookupClusterHosts [zookeeper]")
	}

	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx := errgroup.WithContext(ctx)

	log.Infof("startup with kafka hosts %v", kServers)
	log.Infof("startup with zookeeper hosts %v", zServers)

	for _, fn := range funcs {
		grp.Go(fn(ctx, kServers, zServers))
	}

	waiter := make(chan os.Signal, 1)
	signal.Notify(waiter, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-waiter:
	case <-ctx.Done():
	}

	cancel()
	if err := grp.Wait(); err != nil {
		return errors.Annotate(err, "Wait")
	}

	log.Info("dispatcher finished")
	return nil
}

func handleEntityMessages(ctx goka.Context, msg interface{}, actions ...EventAction) error {
	m, ok := msg.(*NodeContext)
	if !ok {
		return errors.Errorf("invalid message type %+v", msg)
	}

	for _, action := range actions {
		if err := action.ApplyMessage(ctx, m); err != nil {
			return errors.Annotate(err, "ApplyMessage")
		}
	}

	return nil
}

func CreateInputEventConsumer(group goka.Group, inputStream, outputStream goka.Stream, actions ...EventAction) DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(NodeContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleEntityMessages(ctx, msg, actions...); err != nil {
						log.Error(errors.Annotate(err, "handleEntityMessages"))
					}
				}), goka.Output(outputStream, new(HubMessageCodec)),
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

func CreateHubConsumer(group goka.Group, inputStream, outputStream goka.Stream, cb goka.ProcessCallback) DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(HubMessageCodec), cb),
				goka.Output(outputStream, new(HubMessageCodec)),
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

func SetLogger(logger logrus.FieldLogger) {
	log = logger
}

func LookupClusterHosts(host string, port int, params ...string) ([]string, error) {
	ips, err := DNSLookupIP(host, 50)
	if err != nil {
		return nil, errors.Annotate(err, "DNSLookupIP")
	}

	res := []string{}
	for _, ip := range ips {
		res = append(res, fmt.Sprintf(
			"%s:%d%s", ip, port, strings.Join(params, ""),
		))
	}

	return res, nil
}
