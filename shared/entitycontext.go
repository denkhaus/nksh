package shared

type EntityContext struct {
	NodeID  int64
	Context map[string]Properties
}

func (p *EntityContext) Append(key string, prop Properties) {

}
