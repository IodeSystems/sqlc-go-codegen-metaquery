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
		terms, pos, err := parseInternal(s)
		if err != nil {
			return Parsed{}, err
		}
		if pos < 0 {
			return Parsed{Search: s, Terms: terms}, nil
		}
		r := []rune(s)
		if pos > len(r) {
			pos = len(r)
		}
		if pos < 0 {
			pos = 0
		}
		s = string(r[:pos]) + "\\" + string(r[pos:])
	}
	return Parsed{}, fmt.Errorf("search: unwanted-token recoveries exhausted (>%d) for %q", maxRecover, query)
}

// unwanted is panicked by the lexer error listener / parser error strategy at
// the character position that needs escaping. fatal is panicked for
// unrecoverable parse errors.
type unwanted struct{ pos int }
type fatal struct{ msg string }

// parseInternal runs one parse pass. It returns (terms, -1, nil) on success,
// (nil, pos, nil) when an unwanted token at pos should be escaped and retried,
// or (nil, 0, err) on a fatal parse error.
func parseInternal(input string) (terms []Term, unwantedPos int, parseErr error) {
	unwantedPos = -1
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case unwanted:
				unwantedPos = e.pos
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
	return l.terms, -1, nil
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

func (s *errorStrategy) ReportError(recognizer antlr.Parser, e antlr.RecognitionException) {
	switch ex := e.(type) {
	case *antlr.InputMisMatchException:
		panic(unwanted{recognizer.GetCurrentToken().GetStop()})
	case *antlr.NoViableAltException:
		panic(unwanted{ex.GetOffendingToken().GetStart() - 1})
	default:
		tok := ""
		if e.GetOffendingToken() != nil {
			tok = e.GetOffendingToken().GetText()
		}
		panic(fatal{msg: "error parsing search at: " + tok})
	}
}
