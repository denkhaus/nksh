package hub

import "github.com/sirupsen/logrus"

var (
	log logrus.FieldLogger = logrus.New().WithField("package", "hub")
)
