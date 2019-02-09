package shared

import "fmt"

type Properties map[string]interface{}

func (p Properties) MustGet(field string) interface{} {
	if val, ok := p[field]; ok {
		return val
	}
	panic(fmt.Sprintf("Properties:MustGet: field %s undefined", field))
}

type Operation string

var (
	CreatedOperation = Operation("created")
	UpdatedOperation = Operation("updated")
	DeletedOperation = Operation("deleted")
)
