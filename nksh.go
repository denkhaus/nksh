package nksh

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/denkhaus/nksh/shared"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	log logrus.FieldLogger = logrus.New().WithField("package", "nksh")
)

func Startup(kafkaHost, zookeeperHost string, funcs ...shared.DispatcherFunc) error {
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

func SetLogger(logger logrus.FieldLogger) {
	log = logger
}
