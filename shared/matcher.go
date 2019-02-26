package shared

type MatcherFunc func() (bool, func() bool)

type Matcher struct {
	funcs []MatcherFunc
}

func (p *Matcher) Eval() bool {
	var valids, results = 0, 0
	for _, fn := range p.funcs {
		if valid, test := fn(); valid {
			valids++
			if test() {
				results++
			}
		}
	}

	if len(p.funcs) > 0 {
		return valids == results
	}

	return false
}

func NewMatcher(fns ...MatcherFunc) *Matcher {
	m := Matcher{
		funcs: fns,
	}
	return &m
}
