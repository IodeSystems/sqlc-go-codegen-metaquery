package search

import (
	"strings"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

// cond is an intermediate predicate: SQL using ? placeholders plus its bound
// args. The final composite is handed to the Builder as one Filter{Expr,Args},
// whose ? placeholders the Builder renumbers ($N / ?N) at Build time.
type cond struct {
	sql  string
	args []any
}

// compile turns parsed terms into a single composite Filter (or nil when the
// search contributes no predicate), driven by the column metadata + Config.
// Port of dataset's SearchConditionFactory.search.
func compile(terms []Term, cols []metaquery.Column, cfg Config, d metaquery.Dialect) (*metaquery.Filter, error) {
	r := newResolver(cols, cfg, d)

	var acc *cond
	for _, term := range terms {
		p, targeted := r.resolveTarget(term.Target)

		var tc *cond
		for _, val := range term.Values {
			var vc *cond
			if targeted {
				c, ok, err := p.run(val.Value)
				if err != nil {
					return nil, err
				}
				if ok {
					vc = c
				}
			} else {
				// Unqualified (or unknown target): OR over all global columns.
				var gc *cond
				for _, gp := range r.globals {
					c, ok, err := gp.run(val.Value)
					if err != nil {
						return nil, err
					}
					if ok {
						gc = combine(gc, c, Or)
					}
				}
				vc = gc
			}
			// Honor value-level negation in both branches (dataset honors it
			// only for global terms; we extend it to targeted for least surprise).
			if val.Negated {
				vc = negate(vc)
			}
			tc = combine(tc, vc, val.Conj)
		}

		if term.Negated {
			tc = negate(tc)
		}
		acc = combine(acc, tc, term.Conj)
	}

	if acc == nil {
		return nil, nil
	}
	return &metaquery.Filter{Expr: acc.sql, Args: acc.args}, nil
}

// resolver holds compiled providers keyed by lowercased target name/alias, plus
// the ordered set of providers participating in unqualified terms.
type resolver struct {
	targets map[string]provider
	globals []provider
}

type provider struct {
	col Col // zero Col for Named (virtual) searches
	fn  Fn
}

// run evaluates the provider for one value, returning (cond, found, err).
// found=false means the value contributed nothing (Fn returned nil filter).
func (p provider) run(value string) (*cond, bool, error) {
	f, err := p.fn(p.col, value)
	if err != nil {
		return nil, false, err
	}
	if f == nil {
		return nil, false, nil
	}
	c := filterToCond(*f)
	return &c, true, nil
}

func newResolver(cols []metaquery.Column, cfg Config, d metaquery.Dialect) *resolver {
	r := &resolver{targets: make(map[string]provider)}

	for _, col := range cols {
		fc := cfg.Fields[col.Name]
		if fc.Disable {
			continue
		}
		kind := metaquery.ValueKind(col.GoType)
		fn := fc.Search
		if fn == nil {
			fn = defaultFn(kind)
		}
		if fn == nil {
			continue // unsearchable kind, no override
		}
		p := provider{col: Col{Column: col, dialect: d}, fn: fn}
		r.targets[strings.ToLower(col.Name)] = p
		for _, a := range fc.Aliases {
			r.targets[strings.ToLower(a)] = p
		}

		global := defaultGlobal(kind)
		switch fc.Scope {
		case ScopeGlobal:
			global = true
		case ScopeTargeted:
			global = false
		}
		if global {
			r.globals = append(r.globals, p)
		}
	}

	for name, n := range cfg.Named {
		if n.Search == nil {
			continue
		}
		p := provider{fn: n.Search}
		r.targets[strings.ToLower(name)] = p
		if n.Global {
			r.globals = append(r.globals, p)
		}
	}
	return r
}

// resolveTarget returns the provider for a term's target. An empty or
// unresolved target yields targeted=false, so the value is searched globally
// (matching dataset, which drops an unknown target and searches its value).
func (r *resolver) resolveTarget(target string) (provider, bool) {
	if target == "" {
		return provider{}, false
	}
	if p, ok := r.targets[strings.ToLower(target)]; ok {
		return p, true
	}
	return provider{}, false
}

func filterToCond(f metaquery.Filter) cond {
	if f.IsExpr() {
		return cond{sql: f.Expr, args: append([]any(nil), f.Args...)}
	}
	op := f.Op
	if op == "" {
		op = metaquery.OpEq
	}
	if op == metaquery.OpIsNull || op == metaquery.OpIsNotNull {
		return cond{sql: quoteIdent(f.Column) + " " + string(op)}
	}
	return cond{sql: quoteIdent(f.Column) + " " + string(op) + " ?", args: []any{f.Value}}
}

func combine(a, b *cond, op Conjunction) *cond {
	if b == nil {
		return a
	}
	if a == nil {
		return b
	}
	sep := " AND "
	if op == Or {
		sep = " OR "
	}
	args := make([]any, 0, len(a.args)+len(b.args))
	args = append(args, a.args...)
	args = append(args, b.args...)
	return &cond{sql: "(" + a.sql + sep + b.sql + ")", args: args}
}

func negate(c *cond) *cond {
	if c == nil {
		return nil
	}
	return &cond{sql: "NOT (" + c.sql + ")", args: c.args}
}
