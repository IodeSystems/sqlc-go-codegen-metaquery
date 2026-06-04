// Code generated from /home/nthalk/local/src/iodesystems/sqlc-go-codegen-metaquery/metaquery/search/grammar/DataSetSearchParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package search // DataSetSearchParser
import "github.com/antlr4-go/antlr/v4"

// DataSetSearchParserListener is a complete listener for a parse tree produced by DataSetSearchParser.
type DataSetSearchParserListener interface {
	antlr.ParseTreeListener

	// EnterSearch is called when entering the search production.
	EnterSearch(c *SearchContext)

	// EnterSimpleTerm is called when entering the simpleTerm production.
	EnterSimpleTerm(c *SimpleTermContext)

	// EnterAndTerm is called when entering the andTerm production.
	EnterAndTerm(c *AndTermContext)

	// EnterOrTerm is called when entering the orTerm production.
	EnterOrTerm(c *OrTermContext)

	// EnterTerm is called when entering the term production.
	EnterTerm(c *TermContext)

	// EnterTermTarget is called when entering the termTarget production.
	EnterTermTarget(c *TermTargetContext)

	// EnterTermValueGroup is called when entering the termValueGroup production.
	EnterTermValueGroup(c *TermValueGroupContext)

	// EnterSimpleValue is called when entering the simpleValue production.
	EnterSimpleValue(c *SimpleValueContext)

	// EnterTermValue is called when entering the termValue production.
	EnterTermValue(c *TermValueContext)

	// EnterAndValue is called when entering the andValue production.
	EnterAndValue(c *AndValueContext)

	// EnterOrValue is called when entering the orValue production.
	EnterOrValue(c *OrValueContext)

	// EnterUnprotectedOrValue is called when entering the unprotectedOrValue production.
	EnterUnprotectedOrValue(c *UnprotectedOrValueContext)

	// ExitSearch is called when exiting the search production.
	ExitSearch(c *SearchContext)

	// ExitSimpleTerm is called when exiting the simpleTerm production.
	ExitSimpleTerm(c *SimpleTermContext)

	// ExitAndTerm is called when exiting the andTerm production.
	ExitAndTerm(c *AndTermContext)

	// ExitOrTerm is called when exiting the orTerm production.
	ExitOrTerm(c *OrTermContext)

	// ExitTerm is called when exiting the term production.
	ExitTerm(c *TermContext)

	// ExitTermTarget is called when exiting the termTarget production.
	ExitTermTarget(c *TermTargetContext)

	// ExitTermValueGroup is called when exiting the termValueGroup production.
	ExitTermValueGroup(c *TermValueGroupContext)

	// ExitSimpleValue is called when exiting the simpleValue production.
	ExitSimpleValue(c *SimpleValueContext)

	// ExitTermValue is called when exiting the termValue production.
	ExitTermValue(c *TermValueContext)

	// ExitAndValue is called when exiting the andValue production.
	ExitAndValue(c *AndValueContext)

	// ExitOrValue is called when exiting the orValue production.
	ExitOrValue(c *OrValueContext)

	// ExitUnprotectedOrValue is called when exiting the unprotectedOrValue production.
	ExitUnprotectedOrValue(c *UnprotectedOrValueContext)
}
