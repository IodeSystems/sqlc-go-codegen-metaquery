package search

import "strings"

// listener walks the parse tree and accumulates Terms, a faithful port of
// dataset's SearchParser.Listener.
type listener struct {
	*BaseDataSetSearchParserListener
	terms   []Term
	current *Term // points into terms; values appended through it
}

func (l *listener) startTerm(t ITermContext, conj Conjunction) {
	l.terms = append(l.terms, Term{
		Target:  termTarget(t),
		Conj:    conj,
		Negated: t.NEGATE() != nil,
	})
	l.current = &l.terms[len(l.terms)-1]
}

func (l *listener) EnterSimpleTerm(ctx *SimpleTermContext) { l.startTerm(ctx.Term(), And) }
func (l *listener) EnterAndTerm(ctx *AndTermContext)       { l.startTerm(ctx.Term(), And) }
func (l *listener) EnterOrTerm(ctx *OrTermContext)         { l.startTerm(ctx.Term(), Or) }

func (l *listener) EnterSimpleValue(ctx *SimpleValueContext) { l.addValue(ctx.TermValue(), And) }
func (l *listener) EnterAndValue(ctx *AndValueContext)       { l.addValue(ctx.TermValue(), And) }
func (l *listener) EnterOrValue(ctx *OrValueContext)         { l.addValue(ctx.TermValue(), Or) }
func (l *listener) EnterUnprotectedOrValue(ctx *UnprotectedOrValueContext) {
	l.addValue(ctx.TermValue(), Or)
}

func (l *listener) addValue(ctx ITermValueContext, conj Conjunction) {
	val, ok := extractValue(ctx)
	if !ok {
		// No ANY/STRING/ESCAPED_CHAR child. A bare NEGATE is a literal "!".
		if ctx.NEGATE() != nil {
			l.current.Values = append(l.current.Values, TermValue{Conj: conj, Value: ctx.GetText()})
		}
		return
	}
	l.current.Values = append(l.current.Values, TermValue{
		Conj:    conj,
		Value:   val,
		Negated: ctx.NEGATE() != nil,
	})
}

func termTarget(t ITermContext) string {
	tt := t.TermTarget()
	if tt == nil {
		return ""
	}
	return tt.GetText()
}

func extractValue(ctx ITermValueContext) (string, bool) {
	switch {
	case ctx.ANY() != nil:
		return unescapeValue(ctx.ANY().GetText()), true
	case ctx.ESCAPED_CHAR() != nil:
		return unescapeValue(ctx.ESCAPED_CHAR().GetText()), true
	case ctx.STRING() != nil:
		return unescapeValue(ctx.STRING().GetText()), true
	}
	return "", false
}

// unescapeValue strips a single layer of wrapping quotes/parens, otherwise
// resolves backslash escapes — matching dataset's extractValue(String).
func unescapeValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return v
	}
	if wrapped(v, '(', ')') || wrapped(v, '"', '"') || wrapped(v, '\'', '\'') {
		return v[1 : len(v)-1]
	}
	r := strings.NewReplacer(
		`\'`, `'`,
		`\ `, ` `,
		`\"`, `"`,
		`\:`, `:`,
		`\(`, `(`,
		`\)`, `)`,
		`\!`, `!`,
		`\\`, `\`,
	)
	return r.Replace(v)
}

func wrapped(s string, open, close byte) bool {
	return len(s) >= 2 && s[0] == open && s[len(s)-1] == close
}
