package search

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// maxRecover bounds the auto-escape retry loop, matching dataset's 1..1000.
const maxRecover = 1000

// Parse parses a search string into terms, auto-escaping offending characters
// (by inserting a backslash before them) and retrying, up to maxRecover times.
// The returned Parsed.Search is the normalized (possibly escaped) string, which
// callers surface as DataSetResponse.searchRendered.
//
// Faithful port of dataset's SearchParser.parse.
func Parse(query string) (Parsed, error) {
	s := strings.TrimSpace(query)
	for range maxRecover {
		terms, pos, escape, err := parseInternal(s)
		if err != nil {
			return Parsed{}, err
		}
		if !escape {
			return Parsed{Search: s, Terms: terms}, nil
		}
		// Clamp like Kotlin's take/drop, which coerce out-of-range indices: a
		// NoViableAlt at offset 0 yields pos -1 (offendingToken.startIndex-1),
		// meaning "escape at the start".
		r := []rune(s)
		pos = min(max(pos, 0), len(r))
		s = string(r[:pos]) + "\\" + string(r[pos:])
	}
	return Parsed{}, fmt.Errorf("search: unwanted-token recoveries exhausted (>%d) for %q", maxRecover, query)
}

// unwanted is panicked by the lexer error listener / parser error strategy at
// the character position that needs escaping. fatal is panicked for
// unrecoverable parse errors.
type unwanted struct{ pos int }
type fatal struct{ msg string }

// parseInternal runs one parse pass. It returns (terms, 0, false, nil) on
// success, (nil, pos, true, nil) when an unwanted token at pos should be escaped
// and retried, or (nil, 0, false, err) on a fatal parse error. The escape bool
// (not pos sign) is the signal — an escape position can legitimately be -1.
func parseInternal(input string) (terms []Term, unwantedPos int, escape bool, parseErr error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case unwanted:
				unwantedPos, escape = e.pos, true
			case fatal:
				parseErr = fmt.Errorf("search: %s", e.msg)
			default:
				panic(r)
			}
		}
	}()

	lexer := NewDataSetSearchLexer(antlr.NewInputStream(input))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&lexerErrorListener{})

	p := NewDataSetSearchParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	p.RemoveErrorListeners()
	p.SetErrorHandler(&errorStrategy{antlr.NewDefaultErrorStrategy()})

	l := &listener{}
	antlr.ParseTreeWalkerDefault.Walk(l, p.Search())
	return l.terms, 0, false, nil
}

// lexerErrorListener throws (panics) at the offending character position, like
// dataset's anonymous BaseErrorListener on the lexer.
type lexerErrorListener struct {
	*antlr.DefaultErrorListener
}

func (*lexerErrorListener) SyntaxError(_ antlr.Recognizer, _ any, _, column int, _ string, _ antlr.RecognitionException) {
	panic(unwanted{column})
}

// errorStrategy mirrors dataset's DefaultErrorStrategy overrides: an unwanted
// or mismatched token escapes at the current token's stop index; a no-viable
// alternative escapes just before its offending token.
type errorStrategy struct {
	*antlr.DefaultErrorStrategy
}

func (s *errorStrategy) ReportUnwantedToken(recognizer antlr.Parser) {
	panic(unwanted{recognizer.GetCurrentToken().GetStop()})
}

// RecoverInline mirrors Kotlin's fail-fast behavior. The Go DefaultErrorStrategy
// silently single-token-inserts a missing symbol (its SingleTokenInsertion
// returns a synthetic token without reporting), so inputs like "!" or "!!a"
// would parse with a fabricated token instead of triggering the escape-retry
// loop. Panicking here forces the offending position to be escaped and reparsed,
// matching the Kotlin parser.
func (s *errorStrategy) RecoverInline(recognizer antlr.Parser) antlr.Token {
	tok := recognizer.GetCurrentToken()
	if tok == nil {
		panic(fatal{msg: "recover inline"})
	}
	panic(unwanted{tok.GetStop()})
}

func (s *errorStrategy) ReportError(recognizer antlr.Parser, e antlr.RecognitionException) {
	switch ex := e.(type) {
	case *antlr.InputMisMatchException:
		panic(unwanted{recognizer.GetCurrentToken().GetStop()})
	case *antlr.NoViableAltException:
		// Kotlin reads ex.offendingToken.startIndex-1; the Go runtime keeps the
		// NoViableAlt offending token in an unexported field, so GetOffendingToken
		// can be nil (e.g. a lone "!" failing at EOF). Fall back to the current
		// token's start, which is where the alternative got stuck.
		tok := ex.GetOffendingToken()
		if tok == nil {
			tok = recognizer.GetCurrentToken()
		}
		if tok == nil {
			panic(fatal{msg: "no viable alternative"})
		}
		panic(unwanted{tok.GetStart() - 1})
	default:
		tok := ""
		if e.GetOffendingToken() != nil {
			tok = e.GetOffendingToken().GetText()
		}
		panic(fatal{msg: "error parsing search at: " + tok})
	}
}
