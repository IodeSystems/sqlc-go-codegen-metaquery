// Code generated from /home/nthalk/local/src/iodesystems/sqlc-go-codegen-metaquery/metaquery/search/grammar/DataSetSearchLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package search

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type DataSetSearchLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var DataSetSearchLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func datasetsearchlexerLexerInit() {
	staticData := &DataSetSearchLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE", "ESCAPED_MODE",
	}
	staticData.LiteralNames = []string{
		"", "", "'\\'", "'!'", "", "", "':'", "','", "'('", "')'",
	}
	staticData.SymbolicNames = []string{
		"", "ESCAPED_CHAR", "ESCAPE", "NEGATE", "ANY", "STRING", "TARGET_SEPARATOR",
		"TERM_OR", "TERM_GROUP_START", "TERM_GROUP_END", "WS", "ESCAPED",
	}
	staticData.RuleNames = []string{
		"ESCAPED_CHAR", "ESCAPE", "NEGATE", "ANY", "STRING", "TARGET_SEPARATOR",
		"TERM_OR", "TERM_GROUP_START", "TERM_GROUP_END", "WS", "ESCAPED",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 11, 83, 6, -1, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7,
		3, 2, 4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7,
		9, 2, 10, 7, 10, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2,
		1, 3, 1, 3, 4, 3, 36, 8, 3, 11, 3, 12, 3, 37, 1, 3, 1, 3, 5, 3, 42, 8,
		3, 10, 3, 12, 3, 45, 9, 3, 1, 4, 1, 4, 1, 4, 5, 4, 50, 8, 4, 10, 4, 12,
		4, 53, 9, 4, 1, 4, 1, 4, 1, 4, 1, 4, 5, 4, 59, 8, 4, 10, 4, 12, 4, 62,
		9, 4, 1, 4, 3, 4, 65, 8, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1, 8, 1,
		8, 1, 9, 4, 9, 76, 8, 9, 11, 9, 12, 9, 77, 1, 10, 1, 10, 1, 10, 1, 10,
		0, 0, 11, 2, 1, 4, 2, 6, 3, 8, 4, 10, 5, 12, 6, 14, 7, 16, 8, 18, 9, 20,
		10, 22, 11, 2, 0, 1, 5, 5, 0, 32, 34, 39, 41, 44, 44, 58, 58, 92, 92, 6,
		0, 32, 32, 34, 34, 39, 41, 44, 44, 58, 58, 92, 92, 2, 0, 34, 34, 92, 92,
		2, 0, 39, 39, 92, 92, 3, 0, 9, 10, 13, 13, 32, 32, 91, 0, 2, 1, 0, 0, 0,
		0, 4, 1, 0, 0, 0, 0, 6, 1, 0, 0, 0, 0, 8, 1, 0, 0, 0, 0, 10, 1, 0, 0, 0,
		0, 12, 1, 0, 0, 0, 0, 14, 1, 0, 0, 0, 0, 16, 1, 0, 0, 0, 0, 18, 1, 0, 0,
		0, 0, 20, 1, 0, 0, 0, 1, 22, 1, 0, 0, 0, 2, 24, 1, 0, 0, 0, 4, 27, 1, 0,
		0, 0, 6, 31, 1, 0, 0, 0, 8, 35, 1, 0, 0, 0, 10, 64, 1, 0, 0, 0, 12, 66,
		1, 0, 0, 0, 14, 68, 1, 0, 0, 0, 16, 70, 1, 0, 0, 0, 18, 72, 1, 0, 0, 0,
		20, 75, 1, 0, 0, 0, 22, 79, 1, 0, 0, 0, 24, 25, 3, 4, 1, 0, 25, 26, 3,
		22, 10, 0, 26, 3, 1, 0, 0, 0, 27, 28, 5, 92, 0, 0, 28, 29, 1, 0, 0, 0,
		29, 30, 6, 1, 0, 0, 30, 5, 1, 0, 0, 0, 31, 32, 5, 33, 0, 0, 32, 7, 1, 0,
		0, 0, 33, 36, 8, 0, 0, 0, 34, 36, 3, 2, 0, 0, 35, 33, 1, 0, 0, 0, 35, 34,
		1, 0, 0, 0, 36, 37, 1, 0, 0, 0, 37, 35, 1, 0, 0, 0, 37, 38, 1, 0, 0, 0,
		38, 43, 1, 0, 0, 0, 39, 42, 8, 1, 0, 0, 40, 42, 3, 2, 0, 0, 41, 39, 1,
		0, 0, 0, 41, 40, 1, 0, 0, 0, 42, 45, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 43,
		44, 1, 0, 0, 0, 44, 9, 1, 0, 0, 0, 45, 43, 1, 0, 0, 0, 46, 51, 5, 34, 0,
		0, 47, 50, 8, 2, 0, 0, 48, 50, 3, 2, 0, 0, 49, 47, 1, 0, 0, 0, 49, 48,
		1, 0, 0, 0, 50, 53, 1, 0, 0, 0, 51, 49, 1, 0, 0, 0, 51, 52, 1, 0, 0, 0,
		52, 54, 1, 0, 0, 0, 53, 51, 1, 0, 0, 0, 54, 65, 5, 34, 0, 0, 55, 60, 5,
		39, 0, 0, 56, 59, 8, 3, 0, 0, 57, 59, 3, 2, 0, 0, 58, 56, 1, 0, 0, 0, 58,
		57, 1, 0, 0, 0, 59, 62, 1, 0, 0, 0, 60, 58, 1, 0, 0, 0, 60, 61, 1, 0, 0,
		0, 61, 63, 1, 0, 0, 0, 62, 60, 1, 0, 0, 0, 63, 65, 5, 39, 0, 0, 64, 46,
		1, 0, 0, 0, 64, 55, 1, 0, 0, 0, 65, 11, 1, 0, 0, 0, 66, 67, 5, 58, 0, 0,
		67, 13, 1, 0, 0, 0, 68, 69, 5, 44, 0, 0, 69, 15, 1, 0, 0, 0, 70, 71, 5,
		40, 0, 0, 71, 17, 1, 0, 0, 0, 72, 73, 5, 41, 0, 0, 73, 19, 1, 0, 0, 0,
		74, 76, 7, 4, 0, 0, 75, 74, 1, 0, 0, 0, 76, 77, 1, 0, 0, 0, 77, 75, 1,
		0, 0, 0, 77, 78, 1, 0, 0, 0, 78, 21, 1, 0, 0, 0, 79, 80, 9, 0, 0, 0, 80,
		81, 1, 0, 0, 0, 81, 82, 6, 10, 1, 0, 82, 23, 1, 0, 0, 0, 12, 0, 1, 35,
		37, 41, 43, 49, 51, 58, 60, 64, 77, 2, 5, 1, 0, 4, 0, 0,
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

// DataSetSearchLexerInit initializes any static state used to implement DataSetSearchLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewDataSetSearchLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func DataSetSearchLexerInit() {
	staticData := &DataSetSearchLexerLexerStaticData
	staticData.once.Do(datasetsearchlexerLexerInit)
}

// NewDataSetSearchLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewDataSetSearchLexer(input antlr.CharStream) *DataSetSearchLexer {
	DataSetSearchLexerInit()
	l := new(DataSetSearchLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &DataSetSearchLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "DataSetSearchLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// DataSetSearchLexer tokens.
const (
	DataSetSearchLexerESCAPED_CHAR     = 1
	DataSetSearchLexerESCAPE           = 2
	DataSetSearchLexerNEGATE           = 3
	DataSetSearchLexerANY              = 4
	DataSetSearchLexerSTRING           = 5
	DataSetSearchLexerTARGET_SEPARATOR = 6
	DataSetSearchLexerTERM_OR          = 7
	DataSetSearchLexerTERM_GROUP_START = 8
	DataSetSearchLexerTERM_GROUP_END   = 9
	DataSetSearchLexerWS               = 10
	DataSetSearchLexerESCAPED          = 11
)

// DataSetSearchLexerESCAPED_MODE is the DataSetSearchLexer mode.
const DataSetSearchLexerESCAPED_MODE = 1
