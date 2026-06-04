# DataSet search grammar (vendored)

These `.g4` files are **copies**, kept byte-identical to the single source of
truth in the sibling `dataset` project:

    ../dataset/src/main/antlr4/DataSetSearchLexer.g4
    ../dataset/src/main/antlr4/DataSetSearchParser.g4

One grammar, two ANTLR targets (Kotlin in `dataset`, Go here) → guaranteed parse
fidelity. Do not hand-edit the grammar here; edit it in `dataset`, copy it over,
then regenerate the Go parser:

    make antlr   # runs scripts/antlr.sh (ANTLR 4.13.2)

To check for drift:

    diff DataSetSearchLexer.g4  ../../../../dataset/src/main/antlr4/DataSetSearchLexer.g4
    diff DataSetSearchParser.g4 ../../../../dataset/src/main/antlr4/DataSetSearchParser.g4

The generated `*.go` files in the parent directory carry a
`// Code generated ... DO NOT EDIT.` header and must never be edited by hand.
