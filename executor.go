package nksh

import (
	"time"

	"github.com/denkhaus/nksh/hub"
	"github.com/denkhaus/nksh/shared"

	"github.com/juju/errors"
	"github.com/lovoo/goka"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

var (
	ErrEmptyOperationResult   = errors.New("empty operation result")
	ErrInvalidOperationResult = errors.New("invalid operation result")
)

type Executor struct {
	Driver  neo4j.Driver
	Context goka.Context
}

func NewExecutor(ctx goka.Context, driver neo4j.Driver) *Executor {
	ex := Executor{
		Driver:  driver,
		Context: ctx,
	}

	return &ex
}

func (p *Executor) newSession() (neo4j.Session, error) {
	session, err := p.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, errors.Annotate(err, "Session")
	}

	return session, nil
}

func (p *Executor) enumerateSuperOrdinates(

	senderID int64,
	enumerate func(id int64, labels []interface{}) error,

) error {

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
			"id": senderID,
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

func (p *Executor) NotifySuperOrdinates(

	sender string,
	senderID int64,
	operation shared.Operation,
	props shared.Properties,

) error {
	msg := hub.Context{
		Sender:     sender,
		Operation:  operation,
		SenderID:   senderID,
		Properties: props,
	}

	p.enumerateSuperOrdinates(senderID, func(id int64, labels []interface{}) error {
		for _, l := range labels {
			label := l.(string)
			msg.Receiver = label
			msg.ReceiverID = id
			p.Context.Emit(goka.Stream("Hub"), ComposeKey(label, id), msg)
		}

		return nil
	})

	return nil
}

func (p *Executor) ApplyContext(nodeID int64, ctx map[string]interface{}) error {
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
			"id":         nodeID,
			"modifiedAt": time.Now().UTC(),
			"ctx":        ctx,
		})

	if err != nil {
		return errors.Annotate(err, "Run")
	}

	if result.Next() {
		if res, ok := result.Record().Get("result"); ok {
			if res.(int64) != nodeID {
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
