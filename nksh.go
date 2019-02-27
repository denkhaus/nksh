package nksh

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/denkhaus/nksh/hub"
	"github.com/denkhaus/nksh/event"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/kafka"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	log logrus.FieldLogger = logrus.New().WithField("package", "nksh")
	HubStream                = goka.Stream("Hub") 
)


type DispatcherFunc func(ctx context.Context, kServers, zServers []string) func() error

func ComposeKey(label string, id int64) string {
	return fmt.Sprintf("%s-%d-%s", label, id, RandStringBytes(4))
}

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

func handleInputEvents(ctx goka.Context, msg interface{}, actions ...event.Action) error {
	m, ok := msg.(*event.Context)
	if !ok {
		return errors.Errorf("invalid message type %+v", msg)
	}

	for _, action := range actions {
		handled, err := action.ApplyMessage(ctx, m)
		if err != nil {
			return errors.Annotate(err, "ApplyMessage")
		}

		if !handled {
			log.Warningf("unhandled input msg [no match]: %+v", m)
		}
	}

	return nil
}

func CreateInputConsumerDefaults(label string, actions ...event.Action) DispatcherFunc {
	group := goka.Group(fmt.Sprintf("%s_InputGroup",label))
	inputStream := goka.Stream(fmt.Sprintf("Input2%s",label))
 	return CreateInputConsumer(group, inputStream, HubStream , actions ...) 
}

func CreateInputConsumer(group goka.Group, inputStream, outputStream goka.Stream, actions ...event.Action) DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(event.ContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleInputEvents(ctx, msg, actions...); err != nil {
						log.Error(errors.Annotate(err, "handleInputEvents"))
					}
				}), goka.Output(outputStream, new(event.ContextCodec)),
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

func handleHubEvents(ctx goka.Context, msg interface{}, actions ...hub.Action) error {
	m, ok := msg.(*hub.Context)
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

func CreateHubConsumerDefaults(label string, actions ...hub.Action) DispatcherFunc {
	group := goka.Group(fmt.Sprintf("%s_HubGroup",label))
	outputStream := goka.Stream(fmt.Sprintf("Hub2%s",label))
 	return CreateHubConsumer(group, HubStream ,outputStream, actions ...) 
}

func CreateHubConsumer(group goka.Group, inputStream, outputStream goka.Stream, actions ...hub.Action) DispatcherFunc {
	return func(ctx context.Context, kServers, zServers []string) func() error {
		return func() error {
			g := goka.DefineGroup(group,
				goka.Input(inputStream, new(hub.ContextCodec), func(ctx goka.Context, msg interface{}) {
					if err := handleHubEvents(ctx, msg, actions...); err != nil {
						log.Error(errors.Annotate(err, "handleHubEvents"))
					}
				}),
				goka.Output(outputStream, new(hub.ContextCodec)),
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func ConnectNeo4j(host string) (neo4j.Driver, error) {
	log.Info("connect neo4j")

	user := os.Getenv("NEO4J_USERNAME")
	if user == "" {
		return nil, errors.New("Neo4j username undefined")
	}

	password := os.Getenv("NEO4J_PASSWORD")
	if password == "" {
		return nil, errors.New("Neo4j password undefined")
	}

	driver, err := neo4j.NewDriver(fmt.Sprintf("bolt://%s:7687", host),
		neo4j.BasicAuth(user, password, ""),
	)

	if err != nil {
		return nil, errors.Annotate(err, "NewDriver")
	}

	return driver, nil
}
