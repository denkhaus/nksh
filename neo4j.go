package nksh

import (
	"fmt"
	"os"

	"github.com/denkhaus/nksh/shared"

	"github.com/juju/errors"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func ConnectNeo4j(host string) error {
	log.Info("connect neo4j")

	user := os.Getenv("NEO4J_USERNAME")
	if user == "" {
		return errors.New("Neo4j username undefined")
	}

	password := os.Getenv("NEO4J_PASSWORD")
	if password == "" {
		return errors.New("Neo4j password undefined")
	}

	driver, err := neo4j.NewDriver(fmt.Sprintf("bolt://%s:7687", host),
		neo4j.BasicAuth(user, password, ""),
	)

	if err != nil {
		return errors.Annotate(err, "NewDriver")
	}

	shared.Neo4jDriver = driver
	return nil
}

func Neo4jDriver() neo4j.Driver {
	return shared.Neo4jDriver
}

func CloseNeo4j() {
	if shared.Neo4jDriver != nil {
		shared.Neo4jDriver.Close()
		shared.Neo4jDriver = nil
	}
}
