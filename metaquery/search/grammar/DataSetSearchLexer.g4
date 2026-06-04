lexer grammar DataSetSearchLexer;
ESCAPED_CHAR : ESCAPE ESCAPED;
ESCAPE :
    '\\' -> pushMode(ESCAPED_MODE)
;
NEGATE: '!';
ANY: (
  (~(' ' | ':' | ',' | '(' | ')' | '\\' | '"' | '\'' | '!')
      | ESCAPED_CHAR
   )+
  (~(' ' | ':' | ',' | '(' | ')' | '\\' | '"' | '\'')
      | ESCAPED_CHAR
   )*
);
STRING
    : ('"' (~('"' | '\\') | ESCAPED_CHAR)* '"')
    | ('\'' (~('\'' | '\\') | ESCAPED_CHAR)* '\'')
    ;
TARGET_SEPARATOR: ':';
TERM_OR: ',';

TERM_GROUP_START: '(';
TERM_GROUP_END: ')';
WS: [ \t\r\n]+;

mode ESCAPED_MODE;
ESCAPED: . -> popMode;
