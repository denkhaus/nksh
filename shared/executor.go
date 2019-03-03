package shared

import (
	"time"

	"github.com/juju/errors"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

var (
	ErrEmptyOperationResult   = errors.New("empty operation result")
	ErrInvalidOperationResult = errors.New("invalid operation result")
)

var (
	cypherApplyProperties = CypherQuery(`
		MATCH (p) 
		WHERE ID(p) = $id
		SET p+= $ctx 
		SET p.modifiedAt = $modifiedAt		
	`)
	cypherEnumerateSuperOrdinates = CypherQuery(`
		MATCH (super)-[]->(p) 
		WHERE id(p) = $id 
		RETURN ID(super) as id, labels(super) as labels
	`)
)

type OnRecordFunc func(rec neo4j.Record) error

type Executor struct {
	*HandlerContext
}

func NewExecutor(ctx *HandlerContext) *Executor {
	ex := Executor{
		HandlerContext: ctx,
	}

	return &ex
}

func (p *Executor) newSession() (neo4j.Session, error) {
	session, err := Neo4jDriver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, errors.Annotate(err, "Session")
	}

	return session, nil
}

func (p *Executor) enumerateSuperOrdinates(

	senderID int64,
	enumerate func(id int64, labels []interface{}) error,

) error {

	err := p.Run(cypherEnumerateSuperOrdinates,
		Properties{
			"id": senderID,
		}, func(record neo4j.Record) error {
			if i, ok := record.Get("id"); ok {
				id := i.(int64)
				if l, ok := record.Get("labels"); ok {
					if err := enumerate(id, l.([]interface{})); err != nil {
						return errors.Annotate(err, "enumerate")
					}
				}
			}

			return nil
		})

	if err != nil {
		return errors.Annotate(err, "Run")
	}

	return nil
}

func (p *Executor) BuildEntityContext(nodeID int64) (*EntityContext, error) {
	ctx := EntityContext{
		NodeID: nodeID,
	}

	for entity, query := range p.EntityDescriptor.ContextDef() {
		err := p.Run(query, Properties{
			"id": nodeID,
		}, func(record neo4j.Record) error {
			if res, ok := record.Get("result"); ok {
				ctx.Append(entity, res.(map[string]interface{}))
			}
			return nil
		})
		if err != nil {
			return nil, errors.Annotate(err, "Run")
		}
	}

	return &ctx, nil
}

func (p *Executor) NotifySuperOrdinates(

	sender string,
	senderID int64,
	operation Operation,
	props Properties,

) error {

	p.enumerateSuperOrdinates(senderID, func(id int64, labels []interface{}) error {
		for _, l := range labels {

			msg := &HubContext{
				Sender:     sender,
				Operation:  operation,
				SenderID:   senderID,
				Properties: props,
				Receiver:   l.(string),
				ReceiverID: id,
			}

			log.Infof("%s->%s notify superordinate:%v", sender, msg.Receiver, msg)
			p.GokaContext.Emit(HubStream, ComposeKey(msg.Receiver, id), msg)
		}

		return nil
	})

	return nil
}

func (p *Executor) ApplyProperties(nodeID int64, ctx Properties) error {
	err := p.Run(cypherApplyProperties,
		Properties{
			"id":         nodeID,
			"modifiedAt": time.Now().UTC(),
			"ctx":        ctx,
		}, nil)

	if err != nil {
		return errors.Annotate(err, "Run")
	}

	return nil
}

func (p *Executor) Run(cypher CypherQuery, ctx Properties, onRecord OnRecordFunc) error {
	session, err := p.newSession()
	if err != nil {
		return errors.Annotate(err, "newSession")
	}

	defer session.Close()

	result, err := session.Run(cypher.String(), ctx)
	if err != nil {
		return errors.Annotate(err, "Run")
	}

	if onRecord != nil {
		for result.Next() {
			if err := onRecord(result.Record()); err != nil {
				return errors.Annotate(err, "onRecord")
			}
		}
	}

	if err = result.Err(); err != nil {
		return errors.Annotate(err, "Err")
	}

	return nil
}
