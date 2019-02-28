package event

import "github.com/sirupsen/logrus"

var (
	log logrus.FieldLogger = logrus.New().WithField("package", "event")
)
