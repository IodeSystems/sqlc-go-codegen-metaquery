parser grammar DataSetSearchParser;
options {tokenVocab=DataSetSearchLexer;}

// term, term, term, target:term target:
search: (simpleTerm (andTerm|orTerm)*)? WS* TERM_OR? WS* EOF;
simpleTerm: WS* term WS*;
andTerm: WS+ term;
orTerm: WS* TERM_OR WS* term;
term: (termTarget TARGET_SEPARATOR WS*)? NEGATE? WS* termValueGroup;
termTarget:  ANY | STRING;
termValueGroup
: TERM_GROUP_START WS* (simpleValue WS* (andValue|orValue)* WS*)? TERM_GROUP_END
| simpleValue unprotectedOrValue*;
simpleValue: termValue;
termValue: NEGATE? (STRING  | ANY | ESCAPED_CHAR);
andValue: WS+ termValue;
orValue: WS* TERM_OR WS* termValue;
unprotectedOrValue: TERM_OR termValue;
