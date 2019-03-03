package shared

type EvalFunc func(arg interface{}) bool
type EvalFuncs []EvalFunc

type MatcherFunc func() (bool, EvalFunc)

type Matcher struct {
	funcs []MatcherFunc
}

func (p *Matcher) Eval(arg interface{}) bool {
	var valids, results = 0, 0
	for _, fn := range p.funcs {
		if valid, test := fn(); valid {
			valids++
			if test(arg) {
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

func MatchConditions(conditions ...EvalFunc) MatcherFunc {
	return func() (bool, EvalFunc) {
		return len(conditions) > 0,
			func(arg interface{}) bool {
				funcs := []MatcherFunc{}
				for _, cond := range conditions {
					funcs = append(funcs, func() (bool, EvalFunc) {
						return true, cond
					})
				}

				matcher := NewMatcher(funcs...)
				return matcher.Eval(arg)
			}
	}
}
