package search

// Conjunction is how a term (or a value within a term) combines with the next.
type Conjunction int

const (
	And Conjunction = iota
	Or
)

// Term is one parsed search clause: an optional target ("field:"), a negation
// flag, the conjunction relative to the next term, and one or more values.
//
// Mirrors dataset's com.iodesystems.db.search.model.Term so the two language
// targets stay structurally identical.
type Term struct {
	Target  string // "" = unqualified (global) term
	Negated bool
	Conj    Conjunction // relative to the next term
	Values  []TermValue
}

// TermValue is a single value within a term, with its own negation and the
// conjunction relative to the previous value in the group.
type TermValue struct {
	Conj    Conjunction
	Value   string
	Negated bool
}

// Parsed is the result of parsing a search string: the (possibly
// auto-escaped) normalized string and the ordered terms.
type Parsed struct {
	Search string // normalized input, suitable for DataSetResponse.searchRendered
	Terms  []Term
}
