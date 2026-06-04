// Code generated from /home/nthalk/local/src/iodesystems/sqlc-go-codegen-metaquery/metaquery/search/grammar/DataSetSearchParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package search // DataSetSearchParser
import "github.com/antlr4-go/antlr/v4"

// BaseDataSetSearchParserListener is a complete listener for a parse tree produced by DataSetSearchParser.
type BaseDataSetSearchParserListener struct{}

var _ DataSetSearchParserListener = &BaseDataSetSearchParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseDataSetSearchParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseDataSetSearchParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseDataSetSearchParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseDataSetSearchParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSearch is called when production search is entered.
func (s *BaseDataSetSearchParserListener) EnterSearch(ctx *SearchContext) {}

// ExitSearch is called when production search is exited.
func (s *BaseDataSetSearchParserListener) ExitSearch(ctx *SearchContext) {}

// EnterSimpleTerm is called when production simpleTerm is entered.
func (s *BaseDataSetSearchParserListener) EnterSimpleTerm(ctx *SimpleTermContext) {}

// ExitSimpleTerm is called when production simpleTerm is exited.
func (s *BaseDataSetSearchParserListener) ExitSimpleTerm(ctx *SimpleTermContext) {}

// EnterAndTerm is called when production andTerm is entered.
func (s *BaseDataSetSearchParserListener) EnterAndTerm(ctx *AndTermContext) {}

// ExitAndTerm is called when production andTerm is exited.
func (s *BaseDataSetSearchParserListener) ExitAndTerm(ctx *AndTermContext) {}

// EnterOrTerm is called when production orTerm is entered.
func (s *BaseDataSetSearchParserListener) EnterOrTerm(ctx *OrTermContext) {}

// ExitOrTerm is called when production orTerm is exited.
func (s *BaseDataSetSearchParserListener) ExitOrTerm(ctx *OrTermContext) {}

// EnterTerm is called when production term is entered.
func (s *BaseDataSetSearchParserListener) EnterTerm(ctx *TermContext) {}

// ExitTerm is called when production term is exited.
func (s *BaseDataSetSearchParserListener) ExitTerm(ctx *TermContext) {}

// EnterTermTarget is called when production termTarget is entered.
func (s *BaseDataSetSearchParserListener) EnterTermTarget(ctx *TermTargetContext) {}

// ExitTermTarget is called when production termTarget is exited.
func (s *BaseDataSetSearchParserListener) ExitTermTarget(ctx *TermTargetContext) {}

// EnterTermValueGroup is called when production termValueGroup is entered.
func (s *BaseDataSetSearchParserListener) EnterTermValueGroup(ctx *TermValueGroupContext) {}

// ExitTermValueGroup is called when production termValueGroup is exited.
func (s *BaseDataSetSearchParserListener) ExitTermValueGroup(ctx *TermValueGroupContext) {}

// EnterSimpleValue is called when production simpleValue is entered.
func (s *BaseDataSetSearchParserListener) EnterSimpleValue(ctx *SimpleValueContext) {}

// ExitSimpleValue is called when production simpleValue is exited.
func (s *BaseDataSetSearchParserListener) ExitSimpleValue(ctx *SimpleValueContext) {}

// EnterTermValue is called when production termValue is entered.
func (s *BaseDataSetSearchParserListener) EnterTermValue(ctx *TermValueContext) {}

// ExitTermValue is called when production termValue is exited.
func (s *BaseDataSetSearchParserListener) ExitTermValue(ctx *TermValueContext) {}

// EnterAndValue is called when production andValue is entered.
func (s *BaseDataSetSearchParserListener) EnterAndValue(ctx *AndValueContext) {}

// ExitAndValue is called when production andValue is exited.
func (s *BaseDataSetSearchParserListener) ExitAndValue(ctx *AndValueContext) {}

// EnterOrValue is called when production orValue is entered.
func (s *BaseDataSetSearchParserListener) EnterOrValue(ctx *OrValueContext) {}

// ExitOrValue is called when production orValue is exited.
func (s *BaseDataSetSearchParserListener) ExitOrValue(ctx *OrValueContext) {}

// EnterUnprotectedOrValue is called when production unprotectedOrValue is entered.
func (s *BaseDataSetSearchParserListener) EnterUnprotectedOrValue(ctx *UnprotectedOrValueContext) {}

// ExitUnprotectedOrValue is called when production unprotectedOrValue is exited.
func (s *BaseDataSetSearchParserListener) ExitUnprotectedOrValue(ctx *UnprotectedOrValueContext) {}
