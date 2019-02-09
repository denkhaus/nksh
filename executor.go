package nksh

import (
	"time"

	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

var (
	ErrEmptyOperationResult   = errors.New("empty operation result")
	ErrInvalidOperationResult = errors.New("invalid operation result")
)

type Neo4jExecutor struct {
	Driver    neo4j.Driver
	Context   goka.Context
	NodeLabel string
	NodeID    int64
}

func NewNeo4jExecutor(ctx goka.Context, nodeLabel string, nodeID int64, driver neo4j.Driver) *Neo4jExecutor {
	ex := Neo4jExecutor{
		Driver:    driver,
		NodeID:    nodeID,
		Context:   ctx,
		NodeLabel: nodeLabel,
	}

	return &ex
}

func (p *Neo4jExecutor) newSession() (neo4j.Session, error) {
	session, err := p.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, errors.Annotate(err, "Session")
	}

	return session, nil
}

func (p *Neo4jExecutor) enumerateSuperOrdinates(enumerate func(id int64, labels []interface{}) error) error {
	session, err := p.newSession()
	if err != nil {
		return errors.Annotate(err, "newSession")
	}

	defer session.Close()

	result, err := session.Run(`

		MATCH (super)-[]->(p) 
		WHERE id(p) = $id 
		RETURN ID(super) as id, labels(super) as labels
		`,
		map[string]interface{}{
			"id": p.NodeID,
		})

	if err != nil {
		return errors.Annotate(err, "Run")
	}

	if result.Next() {
		if i, ok := result.Record().Get("id"); ok {
			id := i.(int64)
			if l, ok := result.Record().Get("labels"); ok {
				if err := enumerate(id, l.([]interface{})); err != nil {
					return errors.Annotate(err, "enumerate")
				}
			}
		}
	}

	if err = result.Err(); err != nil {
		return errors.Annotate(err, "Err")
	}

	return nil
}

func (p *Neo4jExecutor) NotifySuperOrdinates(reason string, props Properties) error {
	msg := HubMessage{
		SenderLabel:  p.NodeLabel,
		SenderReason: reason,
		SenderID:     p.NodeID,
		Properties:   props,
	}

	p.enumerateSuperOrdinates(func(id int64, labels []interface{}) error {
		for _, l := range labels {
			label := l.(string)
			msg.ReceiverID = id
			msg.ReceiverLabel = label
			p.Context.Emit(goka.Stream("Hub"), ComposeKey(label, id), msg)
		}

		return nil
	})

	return nil
}

func (p *Neo4jExecutor) ApplyContext(ctx map[string]interface{}) error {
	session, err := p.newSession()
	if err != nil {
		return errors.Annotate(err, "newSession")
	}

	defer session.Close()

	result, err := session.Run(`

		MATCH (p) 
		WHERE ID(p) = $id
		SET p+= $ctx 
		RETURN ID(p) as result

		`,
		map[string]interface{}{
			"id":         p.NodeID,
			"modifiedAt": time.Now().UTC(),
			"ctx":        ctx,
		})

	if err != nil {
		return errors.Annotate(err, "Run")
	}

	if result.Next() {
		if res, ok := result.Record().Get("result"); ok {
			if res.(int64) != p.NodeID {
				return ErrInvalidOperationResult
			}
		} else {
			return ErrEmptyOperationResult
		}
	} else {
		return ErrEmptyOperationResult
	}

	if err = result.Err(); err != nil {
		return errors.Annotate(err, "Err")
	}

	return nil
}
