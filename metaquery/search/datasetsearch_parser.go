// Code generated from /home/nthalk/local/src/iodesystems/sqlc-go-codegen-metaquery/metaquery/search/grammar/DataSetSearchParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package search // DataSetSearchParser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type DataSetSearchParser struct {
	*antlr.BaseParser
}

var DataSetSearchParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func datasetsearchparserParserInit() {
	staticData := &DataSetSearchParserParserStaticData
	staticData.LiteralNames = []string{
		"", "", "'\\'", "'!'", "", "", "':'", "','", "'('", "')'",
	}
	staticData.SymbolicNames = []string{
		"", "ESCAPED_CHAR", "ESCAPE", "NEGATE", "ANY", "STRING", "TARGET_SEPARATOR",
		"TERM_OR", "TERM_GROUP_START", "TERM_GROUP_END", "WS", "ESCAPED",
	}
	staticData.RuleNames = []string{
		"search", "simpleTerm", "andTerm", "orTerm", "term", "termTarget", "termValueGroup",
		"simpleValue", "termValue", "andValue", "orValue", "unprotectedOrValue",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 11, 181, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 1, 0, 1, 0, 1, 0, 5, 0, 28, 8, 0, 10, 0, 12, 0, 31, 9,
		0, 3, 0, 33, 8, 0, 1, 0, 5, 0, 36, 8, 0, 10, 0, 12, 0, 39, 9, 0, 1, 0,
		3, 0, 42, 8, 0, 1, 0, 5, 0, 45, 8, 0, 10, 0, 12, 0, 48, 9, 0, 1, 0, 1,
		0, 1, 1, 5, 1, 53, 8, 1, 10, 1, 12, 1, 56, 9, 1, 1, 1, 1, 1, 5, 1, 60,
		8, 1, 10, 1, 12, 1, 63, 9, 1, 1, 2, 4, 2, 66, 8, 2, 11, 2, 12, 2, 67, 1,
		2, 1, 2, 1, 3, 5, 3, 73, 8, 3, 10, 3, 12, 3, 76, 9, 3, 1, 3, 1, 3, 5, 3,
		80, 8, 3, 10, 3, 12, 3, 83, 9, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 5, 4, 90,
		8, 4, 10, 4, 12, 4, 93, 9, 4, 3, 4, 95, 8, 4, 1, 4, 3, 4, 98, 8, 4, 1,
		4, 5, 4, 101, 8, 4, 10, 4, 12, 4, 104, 9, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1,
		6, 1, 6, 5, 6, 112, 8, 6, 10, 6, 12, 6, 115, 9, 6, 1, 6, 1, 6, 5, 6, 119,
		8, 6, 10, 6, 12, 6, 122, 9, 6, 1, 6, 1, 6, 5, 6, 126, 8, 6, 10, 6, 12,
		6, 129, 9, 6, 1, 6, 5, 6, 132, 8, 6, 10, 6, 12, 6, 135, 9, 6, 3, 6, 137,
		8, 6, 1, 6, 1, 6, 1, 6, 5, 6, 142, 8, 6, 10, 6, 12, 6, 145, 9, 6, 3, 6,
		147, 8, 6, 1, 7, 1, 7, 1, 8, 3, 8, 152, 8, 8, 1, 8, 1, 8, 1, 9, 4, 9, 157,
		8, 9, 11, 9, 12, 9, 158, 1, 9, 1, 9, 1, 10, 5, 10, 164, 8, 10, 10, 10,
		12, 10, 167, 9, 10, 1, 10, 1, 10, 5, 10, 171, 8, 10, 10, 10, 12, 10, 174,
		9, 10, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11, 1, 11, 0, 0, 12, 0, 2, 4, 6,
		8, 10, 12, 14, 16, 18, 20, 22, 0, 2, 1, 0, 4, 5, 2, 0, 1, 1, 4, 5, 195,
		0, 32, 1, 0, 0, 0, 2, 54, 1, 0, 0, 0, 4, 65, 1, 0, 0, 0, 6, 74, 1, 0, 0,
		0, 8, 94, 1, 0, 0, 0, 10, 107, 1, 0, 0, 0, 12, 146, 1, 0, 0, 0, 14, 148,
		1, 0, 0, 0, 16, 151, 1, 0, 0, 0, 18, 156, 1, 0, 0, 0, 20, 165, 1, 0, 0,
		0, 22, 177, 1, 0, 0, 0, 24, 29, 3, 2, 1, 0, 25, 28, 3, 4, 2, 0, 26, 28,
		3, 6, 3, 0, 27, 25, 1, 0, 0, 0, 27, 26, 1, 0, 0, 0, 28, 31, 1, 0, 0, 0,
		29, 27, 1, 0, 0, 0, 29, 30, 1, 0, 0, 0, 30, 33, 1, 0, 0, 0, 31, 29, 1,
		0, 0, 0, 32, 24, 1, 0, 0, 0, 32, 33, 1, 0, 0, 0, 33, 37, 1, 0, 0, 0, 34,
		36, 5, 10, 0, 0, 35, 34, 1, 0, 0, 0, 36, 39, 1, 0, 0, 0, 37, 35, 1, 0,
		0, 0, 37, 38, 1, 0, 0, 0, 38, 41, 1, 0, 0, 0, 39, 37, 1, 0, 0, 0, 40, 42,
		5, 7, 0, 0, 41, 40, 1, 0, 0, 0, 41, 42, 1, 0, 0, 0, 42, 46, 1, 0, 0, 0,
		43, 45, 5, 10, 0, 0, 44, 43, 1, 0, 0, 0, 45, 48, 1, 0, 0, 0, 46, 44, 1,
		0, 0, 0, 46, 47, 1, 0, 0, 0, 47, 49, 1, 0, 0, 0, 48, 46, 1, 0, 0, 0, 49,
		50, 5, 0, 0, 1, 50, 1, 1, 0, 0, 0, 51, 53, 5, 10, 0, 0, 52, 51, 1, 0, 0,
		0, 53, 56, 1, 0, 0, 0, 54, 52, 1, 0, 0, 0, 54, 55, 1, 0, 0, 0, 55, 57,
		1, 0, 0, 0, 56, 54, 1, 0, 0, 0, 57, 61, 3, 8, 4, 0, 58, 60, 5, 10, 0, 0,
		59, 58, 1, 0, 0, 0, 60, 63, 1, 0, 0, 0, 61, 59, 1, 0, 0, 0, 61, 62, 1,
		0, 0, 0, 62, 3, 1, 0, 0, 0, 63, 61, 1, 0, 0, 0, 64, 66, 5, 10, 0, 0, 65,
		64, 1, 0, 0, 0, 66, 67, 1, 0, 0, 0, 67, 65, 1, 0, 0, 0, 67, 68, 1, 0, 0,
		0, 68, 69, 1, 0, 0, 0, 69, 70, 3, 8, 4, 0, 70, 5, 1, 0, 0, 0, 71, 73, 5,
		10, 0, 0, 72, 71, 1, 0, 0, 0, 73, 76, 1, 0, 0, 0, 74, 72, 1, 0, 0, 0, 74,
		75, 1, 0, 0, 0, 75, 77, 1, 0, 0, 0, 76, 74, 1, 0, 0, 0, 77, 81, 5, 7, 0,
		0, 78, 80, 5, 10, 0, 0, 79, 78, 1, 0, 0, 0, 80, 83, 1, 0, 0, 0, 81, 79,
		1, 0, 0, 0, 81, 82, 1, 0, 0, 0, 82, 84, 1, 0, 0, 0, 83, 81, 1, 0, 0, 0,
		84, 85, 3, 8, 4, 0, 85, 7, 1, 0, 0, 0, 86, 87, 3, 10, 5, 0, 87, 91, 5,
		6, 0, 0, 88, 90, 5, 10, 0, 0, 89, 88, 1, 0, 0, 0, 90, 93, 1, 0, 0, 0, 91,
		89, 1, 0, 0, 0, 91, 92, 1, 0, 0, 0, 92, 95, 1, 0, 0, 0, 93, 91, 1, 0, 0,
		0, 94, 86, 1, 0, 0, 0, 94, 95, 1, 0, 0, 0, 95, 97, 1, 0, 0, 0, 96, 98,
		5, 3, 0, 0, 97, 96, 1, 0, 0, 0, 97, 98, 1, 0, 0, 0, 98, 102, 1, 0, 0, 0,
		99, 101, 5, 10, 0, 0, 100, 99, 1, 0, 0, 0, 101, 104, 1, 0, 0, 0, 102, 100,
		1, 0, 0, 0, 102, 103, 1, 0, 0, 0, 103, 105, 1, 0, 0, 0, 104, 102, 1, 0,
		0, 0, 105, 106, 3, 12, 6, 0, 106, 9, 1, 0, 0, 0, 107, 108, 7, 0, 0, 0,
		108, 11, 1, 0, 0, 0, 109, 113, 5, 8, 0, 0, 110, 112, 5, 10, 0, 0, 111,
		110, 1, 0, 0, 0, 112, 115, 1, 0, 0, 0, 113, 111, 1, 0, 0, 0, 113, 114,
		1, 0, 0, 0, 114, 136, 1, 0, 0, 0, 115, 113, 1, 0, 0, 0, 116, 120, 3, 14,
		7, 0, 117, 119, 5, 10, 0, 0, 118, 117, 1, 0, 0, 0, 119, 122, 1, 0, 0, 0,
		120, 118, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0, 121, 127, 1, 0, 0, 0, 122,
		120, 1, 0, 0, 0, 123, 126, 3, 18, 9, 0, 124, 126, 3, 20, 10, 0, 125, 123,
		1, 0, 0, 0, 125, 124, 1, 0, 0, 0, 126, 129, 1, 0, 0, 0, 127, 125, 1, 0,
		0, 0, 127, 128, 1, 0, 0, 0, 128, 133, 1, 0, 0, 0, 129, 127, 1, 0, 0, 0,
		130, 132, 5, 10, 0, 0, 131, 130, 1, 0, 0, 0, 132, 135, 1, 0, 0, 0, 133,
		131, 1, 0, 0, 0, 133, 134, 1, 0, 0, 0, 134, 137, 1, 0, 0, 0, 135, 133,
		1, 0, 0, 0, 136, 116, 1, 0, 0, 0, 136, 137, 1, 0, 0, 0, 137, 138, 1, 0,
		0, 0, 138, 147, 5, 9, 0, 0, 139, 143, 3, 14, 7, 0, 140, 142, 3, 22, 11,
		0, 141, 140, 1, 0, 0, 0, 142, 145, 1, 0, 0, 0, 143, 141, 1, 0, 0, 0, 143,
		144, 1, 0, 0, 0, 144, 147, 1, 0, 0, 0, 145, 143, 1, 0, 0, 0, 146, 109,
		1, 0, 0, 0, 146, 139, 1, 0, 0, 0, 147, 13, 1, 0, 0, 0, 148, 149, 3, 16,
		8, 0, 149, 15, 1, 0, 0, 0, 150, 152, 5, 3, 0, 0, 151, 150, 1, 0, 0, 0,
		151, 152, 1, 0, 0, 0, 152, 153, 1, 0, 0, 0, 153, 154, 7, 1, 0, 0, 154,
		17, 1, 0, 0, 0, 155, 157, 5, 10, 0, 0, 156, 155, 1, 0, 0, 0, 157, 158,
		1, 0, 0, 0, 158, 156, 1, 0, 0, 0, 158, 159, 1, 0, 0, 0, 159, 160, 1, 0,
		0, 0, 160, 161, 3, 16, 8, 0, 161, 19, 1, 0, 0, 0, 162, 164, 5, 10, 0, 0,
		163, 162, 1, 0, 0, 0, 164, 167, 1, 0, 0, 0, 165, 163, 1, 0, 0, 0, 165,
		166, 1, 0, 0, 0, 166, 168, 1, 0, 0, 0, 167, 165, 1, 0, 0, 0, 168, 172,
		5, 7, 0, 0, 169, 171, 5, 10, 0, 0, 170, 169, 1, 0, 0, 0, 171, 174, 1, 0,
		0, 0, 172, 170, 1, 0, 0, 0, 172, 173, 1, 0, 0, 0, 173, 175, 1, 0, 0, 0,
		174, 172, 1, 0, 0, 0, 175, 176, 3, 16, 8, 0, 176, 21, 1, 0, 0, 0, 177,
		178, 5, 7, 0, 0, 178, 179, 3, 16, 8, 0, 179, 23, 1, 0, 0, 0, 27, 27, 29,
		32, 37, 41, 46, 54, 61, 67, 74, 81, 91, 94, 97, 102, 113, 120, 125, 127,
		133, 136, 143, 146, 151, 158, 165, 172,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// DataSetSearchParserInit initializes any static state used to implement DataSetSearchParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewDataSetSearchParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func DataSetSearchParserInit() {
	staticData := &DataSetSearchParserParserStaticData
	staticData.once.Do(datasetsearchparserParserInit)
}

// NewDataSetSearchParser produces a new parser instance for the optional input antlr.TokenStream.
func NewDataSetSearchParser(input antlr.TokenStream) *DataSetSearchParser {
	DataSetSearchParserInit()
	this := new(DataSetSearchParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &DataSetSearchParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "DataSetSearchParser.g4"

	return this
}

// DataSetSearchParser tokens.
const (
	DataSetSearchParserEOF              = antlr.TokenEOF
	DataSetSearchParserESCAPED_CHAR     = 1
	DataSetSearchParserESCAPE           = 2
	DataSetSearchParserNEGATE           = 3
	DataSetSearchParserANY              = 4
	DataSetSearchParserSTRING           = 5
	DataSetSearchParserTARGET_SEPARATOR = 6
	DataSetSearchParserTERM_OR          = 7
	DataSetSearchParserTERM_GROUP_START = 8
	DataSetSearchParserTERM_GROUP_END   = 9
	DataSetSearchParserWS               = 10
	DataSetSearchParserESCAPED          = 11
)

// DataSetSearchParser rules.
const (
	DataSetSearchParserRULE_search             = 0
	DataSetSearchParserRULE_simpleTerm         = 1
	DataSetSearchParserRULE_andTerm            = 2
	DataSetSearchParserRULE_orTerm             = 3
	DataSetSearchParserRULE_term               = 4
	DataSetSearchParserRULE_termTarget         = 5
	DataSetSearchParserRULE_termValueGroup     = 6
	DataSetSearchParserRULE_simpleValue        = 7
	DataSetSearchParserRULE_termValue          = 8
	DataSetSearchParserRULE_andValue           = 9
	DataSetSearchParserRULE_orValue            = 10
	DataSetSearchParserRULE_unprotectedOrValue = 11
)

// ISearchContext is an interface to support dynamic dispatch.
type ISearchContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	SimpleTerm() ISimpleTermContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	TERM_OR() antlr.TerminalNode
	AllAndTerm() []IAndTermContext
	AndTerm(i int) IAndTermContext
	AllOrTerm() []IOrTermContext
	OrTerm(i int) IOrTermContext

	// IsSearchContext differentiates from other interfaces.
	IsSearchContext()
}

type SearchContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchContext() *SearchContext {
	var p = new(SearchContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_search
	return p
}

func InitEmptySearchContext(p *SearchContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_search
}

func (*SearchContext) IsSearchContext() {}

func NewSearchContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchContext {
	var p = new(SearchContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_search

	return p
}

func (s *SearchContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchContext) EOF() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserEOF, 0)
}

func (s *SearchContext) SimpleTerm() ISimpleTermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISimpleTermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISimpleTermContext)
}

func (s *SearchContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *SearchContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *SearchContext) TERM_OR() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTERM_OR, 0)
}

func (s *SearchContext) AllAndTerm() []IAndTermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAndTermContext); ok {
			len++
		}
	}

	tst := make([]IAndTermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAndTermContext); ok {
			tst[i] = t.(IAndTermContext)
			i++
		}
	}

	return tst
}

func (s *SearchContext) AndTerm(i int) IAndTermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAndTermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAndTermContext)
}

func (s *SearchContext) AllOrTerm() []IOrTermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrTermContext); ok {
			len++
		}
	}

	tst := make([]IOrTermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrTermContext); ok {
			tst[i] = t.(IOrTermContext)
			i++
		}
	}

	return tst
}

func (s *SearchContext) OrTerm(i int) IOrTermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrTermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrTermContext)
}

func (s *SearchContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterSearch(s)
	}
}

func (s *SearchContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitSearch(s)
	}
}

func (p *DataSetSearchParser) Search() (localctx ISearchContext) {
	localctx = NewSearchContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, DataSetSearchParserRULE_search)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(32)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(24)
			p.SimpleTerm()
		}
		p.SetState(29)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				p.SetState(27)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}

				switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext()) {
				case 1:
					{
						p.SetState(25)
						p.AndTerm()
					}

				case 2:
					{
						p.SetState(26)
						p.OrTerm()
					}

				case antlr.ATNInvalidAltNumber:
					goto errorExit
				}

			}
			p.SetState(31)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	p.SetState(37)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(34)
				p.Match(DataSetSearchParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(39)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(41)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DataSetSearchParserTERM_OR {
		{
			p.SetState(40)
			p.Match(DataSetSearchParserTERM_OR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(46)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DataSetSearchParserWS {
		{
			p.SetState(43)
			p.Match(DataSetSearchParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(48)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(49)
		p.Match(DataSetSearchParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISimpleTermContext is an interface to support dynamic dispatch.
type ISimpleTermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Term() ITermContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsSimpleTermContext differentiates from other interfaces.
	IsSimpleTermContext()
}

type SimpleTermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimpleTermContext() *SimpleTermContext {
	var p = new(SimpleTermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_simpleTerm
	return p
}

func InitEmptySimpleTermContext(p *SimpleTermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_simpleTerm
}

func (*SimpleTermContext) IsSimpleTermContext() {}

func NewSimpleTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SimpleTermContext {
	var p = new(SimpleTermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_simpleTerm

	return p
}

func (s *SimpleTermContext) GetParser() antlr.Parser { return s.parser }

func (s *SimpleTermContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *SimpleTermContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *SimpleTermContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *SimpleTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SimpleTermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SimpleTermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterSimpleTerm(s)
	}
}

func (s *SimpleTermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitSimpleTerm(s)
	}
}

func (p *DataSetSearchParser) SimpleTerm() (localctx ISimpleTermContext) {
	localctx = NewSimpleTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, DataSetSearchParserRULE_simpleTerm)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(54)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(51)
				p.Match(DataSetSearchParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(56)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	{
		p.SetState(57)
		p.Term()
	}
	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(58)
				p.Match(DataSetSearchParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAndTermContext is an interface to support dynamic dispatch.
type IAndTermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Term() ITermContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsAndTermContext differentiates from other interfaces.
	IsAndTermContext()
}

type AndTermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAndTermContext() *AndTermContext {
	var p = new(AndTermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_andTerm
	return p
}

func InitEmptyAndTermContext(p *AndTermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_andTerm
}

func (*AndTermContext) IsAndTermContext() {}

func NewAndTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AndTermContext {
	var p = new(AndTermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_andTerm

	return p
}

func (s *AndTermContext) GetParser() antlr.Parser { return s.parser }

func (s *AndTermContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *AndTermContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *AndTermContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *AndTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndTermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AndTermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterAndTerm(s)
	}
}

func (s *AndTermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitAndTerm(s)
	}
}

func (p *DataSetSearchParser) AndTerm() (localctx IAndTermContext) {
	localctx = NewAndTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, DataSetSearchParserRULE_andTerm)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(65)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = 1
	for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1:
			{
				p.SetState(64)
				p.Match(DataSetSearchParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(67)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	{
		p.SetState(69)
		p.Term()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrTermContext is an interface to support dynamic dispatch.
type IOrTermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_OR() antlr.TerminalNode
	Term() ITermContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsOrTermContext differentiates from other interfaces.
	IsOrTermContext()
}

type OrTermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrTermContext() *OrTermContext {
	var p = new(OrTermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_orTerm
	return p
}

func InitEmptyOrTermContext(p *OrTermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_orTerm
}

func (*OrTermContext) IsOrTermContext() {}

func NewOrTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrTermContext {
	var p = new(OrTermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_orTerm

	return p
}

func (s *OrTermContext) GetParser() antlr.Parser { return s.parser }

func (s *OrTermContext) TERM_OR() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTERM_OR, 0)
}

func (s *OrTermContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *OrTermContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *OrTermContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *OrTermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrTermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrTermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterOrTerm(s)
	}
}

func (s *OrTermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitOrTerm(s)
	}
}

func (p *DataSetSearchParser) OrTerm() (localctx IOrTermContext) {
	localctx = NewOrTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, DataSetSearchParserRULE_orTerm)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(74)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DataSetSearchParserWS {
		{
			p.SetState(71)
			p.Match(DataSetSearchParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(76)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(77)
		p.Match(DataSetSearchParserTERM_OR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(81)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(78)
				p.Match(DataSetSearchParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(83)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	{
		p.SetState(84)
		p.Term()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITermContext is an interface to support dynamic dispatch.
type ITermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TermValueGroup() ITermValueGroupContext
	TermTarget() ITermTargetContext
	TARGET_SEPARATOR() antlr.TerminalNode
	NEGATE() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsTermContext differentiates from other interfaces.
	IsTermContext()
}

type TermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermContext() *TermContext {
	var p = new(TermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_term
	return p
}

func InitEmptyTermContext(p *TermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_term
}

func (*TermContext) IsTermContext() {}

func NewTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermContext {
	var p = new(TermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_term

	return p
}

func (s *TermContext) GetParser() antlr.Parser { return s.parser }

func (s *TermContext) TermValueGroup() ITermValueGroupContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermValueGroupContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermValueGroupContext)
}

func (s *TermContext) TermTarget() ITermTargetContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermTargetContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermTargetContext)
}

func (s *TermContext) TARGET_SEPARATOR() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTARGET_SEPARATOR, 0)
}

func (s *TermContext) NEGATE() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserNEGATE, 0)
}

func (s *TermContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *TermContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *TermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterTerm(s)
	}
}

func (s *TermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitTerm(s)
	}
}

func (p *DataSetSearchParser) Term() (localctx ITermContext) {
	localctx = NewTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, DataSetSearchParserRULE_term)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(94)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(86)
			p.TermTarget()
		}
		{
			p.SetState(87)
			p.Match(DataSetSearchParserTARGET_SEPARATOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(91)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(88)
					p.Match(DataSetSearchParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(93)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	p.SetState(97)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(96)
			p.Match(DataSetSearchParserNEGATE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	p.SetState(102)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DataSetSearchParserWS {
		{
			p.SetState(99)
			p.Match(DataSetSearchParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(104)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(105)
		p.TermValueGroup()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITermTargetContext is an interface to support dynamic dispatch.
type ITermTargetContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ANY() antlr.TerminalNode
	STRING() antlr.TerminalNode

	// IsTermTargetContext differentiates from other interfaces.
	IsTermTargetContext()
}

type TermTargetContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermTargetContext() *TermTargetContext {
	var p = new(TermTargetContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_termTarget
	return p
}

func InitEmptyTermTargetContext(p *TermTargetContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_termTarget
}

func (*TermTargetContext) IsTermTargetContext() {}

func NewTermTargetContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermTargetContext {
	var p = new(TermTargetContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_termTarget

	return p
}

func (s *TermTargetContext) GetParser() antlr.Parser { return s.parser }

func (s *TermTargetContext) ANY() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserANY, 0)
}

func (s *TermTargetContext) STRING() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserSTRING, 0)
}

func (s *TermTargetContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermTargetContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TermTargetContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterTermTarget(s)
	}
}

func (s *TermTargetContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitTermTarget(s)
	}
}

func (p *DataSetSearchParser) TermTarget() (localctx ITermTargetContext) {
	localctx = NewTermTargetContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, DataSetSearchParserRULE_termTarget)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(107)
		_la = p.GetTokenStream().LA(1)

		if !(_la == DataSetSearchParserANY || _la == DataSetSearchParserSTRING) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITermValueGroupContext is an interface to support dynamic dispatch.
type ITermValueGroupContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_GROUP_START() antlr.TerminalNode
	TERM_GROUP_END() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	SimpleValue() ISimpleValueContext
	AllAndValue() []IAndValueContext
	AndValue(i int) IAndValueContext
	AllOrValue() []IOrValueContext
	OrValue(i int) IOrValueContext
	AllUnprotectedOrValue() []IUnprotectedOrValueContext
	UnprotectedOrValue(i int) IUnprotectedOrValueContext

	// IsTermValueGroupContext differentiates from other interfaces.
	IsTermValueGroupContext()
}

type TermValueGroupContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermValueGroupContext() *TermValueGroupContext {
	var p = new(TermValueGroupContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_termValueGroup
	return p
}

func InitEmptyTermValueGroupContext(p *TermValueGroupContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_termValueGroup
}

func (*TermValueGroupContext) IsTermValueGroupContext() {}

func NewTermValueGroupContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermValueGroupContext {
	var p = new(TermValueGroupContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_termValueGroup

	return p
}

func (s *TermValueGroupContext) GetParser() antlr.Parser { return s.parser }

func (s *TermValueGroupContext) TERM_GROUP_START() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTERM_GROUP_START, 0)
}

func (s *TermValueGroupContext) TERM_GROUP_END() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTERM_GROUP_END, 0)
}

func (s *TermValueGroupContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *TermValueGroupContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *TermValueGroupContext) SimpleValue() ISimpleValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISimpleValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISimpleValueContext)
}

func (s *TermValueGroupContext) AllAndValue() []IAndValueContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAndValueContext); ok {
			len++
		}
	}

	tst := make([]IAndValueContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAndValueContext); ok {
			tst[i] = t.(IAndValueContext)
			i++
		}
	}

	return tst
}

func (s *TermValueGroupContext) AndValue(i int) IAndValueContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAndValueContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAndValueContext)
}

func (s *TermValueGroupContext) AllOrValue() []IOrValueContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrValueContext); ok {
			len++
		}
	}

	tst := make([]IOrValueContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrValueContext); ok {
			tst[i] = t.(IOrValueContext)
			i++
		}
	}

	return tst
}

func (s *TermValueGroupContext) OrValue(i int) IOrValueContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrValueContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrValueContext)
}

func (s *TermValueGroupContext) AllUnprotectedOrValue() []IUnprotectedOrValueContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUnprotectedOrValueContext); ok {
			len++
		}
	}

	tst := make([]IUnprotectedOrValueContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUnprotectedOrValueContext); ok {
			tst[i] = t.(IUnprotectedOrValueContext)
			i++
		}
	}

	return tst
}

func (s *TermValueGroupContext) UnprotectedOrValue(i int) IUnprotectedOrValueContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnprotectedOrValueContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnprotectedOrValueContext)
}

func (s *TermValueGroupContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermValueGroupContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TermValueGroupContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterTermValueGroup(s)
	}
}

func (s *TermValueGroupContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitTermValueGroup(s)
	}
}

func (p *DataSetSearchParser) TermValueGroup() (localctx ITermValueGroupContext) {
	localctx = NewTermValueGroupContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, DataSetSearchParserRULE_termValueGroup)
	var _la int

	var _alt int

	p.SetState(146)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case DataSetSearchParserTERM_GROUP_START:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(109)
			p.Match(DataSetSearchParserTERM_GROUP_START)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(113)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == DataSetSearchParserWS {
			{
				p.SetState(110)
				p.Match(DataSetSearchParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(115)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(136)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&58) != 0 {
			{
				p.SetState(116)
				p.SimpleValue()
			}
			p.SetState(120)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
			for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
				if _alt == 1 {
					{
						p.SetState(117)
						p.Match(DataSetSearchParserWS)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				p.SetState(122)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext())
				if p.HasError() {
					goto errorExit
				}
			}
			p.SetState(127)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
			for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
				if _alt == 1 {
					p.SetState(125)
					p.GetErrorHandler().Sync(p)
					if p.HasError() {
						goto errorExit
					}

					switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 17, p.GetParserRuleContext()) {
					case 1:
						{
							p.SetState(123)
							p.AndValue()
						}

					case 2:
						{
							p.SetState(124)
							p.OrValue()
						}

					case antlr.ATNInvalidAltNumber:
						goto errorExit
					}

				}
				p.SetState(129)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext())
				if p.HasError() {
					goto errorExit
				}
			}
			p.SetState(133)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == DataSetSearchParserWS {
				{
					p.SetState(130)
					p.Match(DataSetSearchParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(135)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(138)
			p.Match(DataSetSearchParserTERM_GROUP_END)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case DataSetSearchParserESCAPED_CHAR, DataSetSearchParserNEGATE, DataSetSearchParserANY, DataSetSearchParserSTRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(139)
			p.SimpleValue()
		}
		p.SetState(143)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(140)
					p.UnprotectedOrValue()
				}

			}
			p.SetState(145)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISimpleValueContext is an interface to support dynamic dispatch.
type ISimpleValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TermValue() ITermValueContext

	// IsSimpleValueContext differentiates from other interfaces.
	IsSimpleValueContext()
}

type SimpleValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySimpleValueContext() *SimpleValueContext {
	var p = new(SimpleValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_simpleValue
	return p
}

func InitEmptySimpleValueContext(p *SimpleValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_simpleValue
}

func (*SimpleValueContext) IsSimpleValueContext() {}

func NewSimpleValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SimpleValueContext {
	var p = new(SimpleValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_simpleValue

	return p
}

func (s *SimpleValueContext) GetParser() antlr.Parser { return s.parser }

func (s *SimpleValueContext) TermValue() ITermValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermValueContext)
}

func (s *SimpleValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SimpleValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SimpleValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterSimpleValue(s)
	}
}

func (s *SimpleValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitSimpleValue(s)
	}
}

func (p *DataSetSearchParser) SimpleValue() (localctx ISimpleValueContext) {
	localctx = NewSimpleValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, DataSetSearchParserRULE_simpleValue)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(148)
		p.TermValue()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITermValueContext is an interface to support dynamic dispatch.
type ITermValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	ANY() antlr.TerminalNode
	ESCAPED_CHAR() antlr.TerminalNode
	NEGATE() antlr.TerminalNode

	// IsTermValueContext differentiates from other interfaces.
	IsTermValueContext()
}

type TermValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermValueContext() *TermValueContext {
	var p = new(TermValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_termValue
	return p
}

func InitEmptyTermValueContext(p *TermValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_termValue
}

func (*TermValueContext) IsTermValueContext() {}

func NewTermValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermValueContext {
	var p = new(TermValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_termValue

	return p
}

func (s *TermValueContext) GetParser() antlr.Parser { return s.parser }

func (s *TermValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserSTRING, 0)
}

func (s *TermValueContext) ANY() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserANY, 0)
}

func (s *TermValueContext) ESCAPED_CHAR() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserESCAPED_CHAR, 0)
}

func (s *TermValueContext) NEGATE() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserNEGATE, 0)
}

func (s *TermValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TermValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterTermValue(s)
	}
}

func (s *TermValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitTermValue(s)
	}
}

func (p *DataSetSearchParser) TermValue() (localctx ITermValueContext) {
	localctx = NewTermValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, DataSetSearchParserRULE_termValue)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(151)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == DataSetSearchParserNEGATE {
		{
			p.SetState(150)
			p.Match(DataSetSearchParserNEGATE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(153)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&50) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAndValueContext is an interface to support dynamic dispatch.
type IAndValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TermValue() ITermValueContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsAndValueContext differentiates from other interfaces.
	IsAndValueContext()
}

type AndValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAndValueContext() *AndValueContext {
	var p = new(AndValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_andValue
	return p
}

func InitEmptyAndValueContext(p *AndValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_andValue
}

func (*AndValueContext) IsAndValueContext() {}

func NewAndValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AndValueContext {
	var p = new(AndValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_andValue

	return p
}

func (s *AndValueContext) GetParser() antlr.Parser { return s.parser }

func (s *AndValueContext) TermValue() ITermValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermValueContext)
}

func (s *AndValueContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *AndValueContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *AndValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AndValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterAndValue(s)
	}
}

func (s *AndValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitAndValue(s)
	}
}

func (p *DataSetSearchParser) AndValue() (localctx IAndValueContext) {
	localctx = NewAndValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, DataSetSearchParserRULE_andValue)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(156)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == DataSetSearchParserWS {
		{
			p.SetState(155)
			p.Match(DataSetSearchParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(158)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(160)
		p.TermValue()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrValueContext is an interface to support dynamic dispatch.
type IOrValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_OR() antlr.TerminalNode
	TermValue() ITermValueContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsOrValueContext differentiates from other interfaces.
	IsOrValueContext()
}

type OrValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrValueContext() *OrValueContext {
	var p = new(OrValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_orValue
	return p
}

func InitEmptyOrValueContext(p *OrValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_orValue
}

func (*OrValueContext) IsOrValueContext() {}

func NewOrValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrValueContext {
	var p = new(OrValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_orValue

	return p
}

func (s *OrValueContext) GetParser() antlr.Parser { return s.parser }

func (s *OrValueContext) TERM_OR() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTERM_OR, 0)
}

func (s *OrValueContext) TermValue() ITermValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermValueContext)
}

func (s *OrValueContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(DataSetSearchParserWS)
}

func (s *OrValueContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserWS, i)
}

func (s *OrValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterOrValue(s)
	}
}

func (s *OrValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitOrValue(s)
	}
}

func (p *DataSetSearchParser) OrValue() (localctx IOrValueContext) {
	localctx = NewOrValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, DataSetSearchParserRULE_orValue)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(165)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DataSetSearchParserWS {
		{
			p.SetState(162)
			p.Match(DataSetSearchParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(167)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(168)
		p.Match(DataSetSearchParserTERM_OR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(172)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == DataSetSearchParserWS {
		{
			p.SetState(169)
			p.Match(DataSetSearchParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(174)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(175)
		p.TermValue()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUnprotectedOrValueContext is an interface to support dynamic dispatch.
type IUnprotectedOrValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TERM_OR() antlr.TerminalNode
	TermValue() ITermValueContext

	// IsUnprotectedOrValueContext differentiates from other interfaces.
	IsUnprotectedOrValueContext()
}

type UnprotectedOrValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnprotectedOrValueContext() *UnprotectedOrValueContext {
	var p = new(UnprotectedOrValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_unprotectedOrValue
	return p
}

func InitEmptyUnprotectedOrValueContext(p *UnprotectedOrValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = DataSetSearchParserRULE_unprotectedOrValue
}

func (*UnprotectedOrValueContext) IsUnprotectedOrValueContext() {}

func NewUnprotectedOrValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnprotectedOrValueContext {
	var p = new(UnprotectedOrValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = DataSetSearchParserRULE_unprotectedOrValue

	return p
}

func (s *UnprotectedOrValueContext) GetParser() antlr.Parser { return s.parser }

func (s *UnprotectedOrValueContext) TERM_OR() antlr.TerminalNode {
	return s.GetToken(DataSetSearchParserTERM_OR, 0)
}

func (s *UnprotectedOrValueContext) TermValue() ITermValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermValueContext)
}

func (s *UnprotectedOrValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnprotectedOrValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnprotectedOrValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.EnterUnprotectedOrValue(s)
	}
}

func (s *UnprotectedOrValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(DataSetSearchParserListener); ok {
		listenerT.ExitUnprotectedOrValue(s)
	}
}

func (p *DataSetSearchParser) UnprotectedOrValue() (localctx IUnprotectedOrValueContext) {
	localctx = NewUnprotectedOrValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, DataSetSearchParserRULE_unprotectedOrValue)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(177)
		p.Match(DataSetSearchParserTERM_OR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(178)
		p.TermValue()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
