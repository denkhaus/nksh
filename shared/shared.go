package shared

import "fmt"

type Properties map[string]interface{}

func (p Properties) MustGet(field string) interface{} {
	if val, ok := p[field]; ok {
		return val
	}
	panic(fmt.Sprintf("Properties:MustGet: field %s undefined", field))
}

func (p Properties) MustString(field string) string {
	if value, ok := p.MustGet(field).(string); ok {
		return value
	}

	panic(fmt.Sprintf("Properties:MustString: field %s not of type string", field))
}

func (p Properties) MustBool(field string) bool {
	if value, ok := p.MustGet(field).(bool); ok {
		return value
	}

	panic(fmt.Sprintf("Properties:MustBool: field %s not of type bool", field))
}

func (p Properties) MustInt64(field string) int64 {
	if value, ok := p.MustGet(field).(int64); ok {
		return value
	}

	panic(fmt.Sprintf("Properties:MustInt64: field %s not of type int64", field))
}

type Operation string

var (
	CreatedOperation = Operation("created")
	UpdatedOperation = Operation("updated")
	DeletedOperation = Operation("deleted")
)
