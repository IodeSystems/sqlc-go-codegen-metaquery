#!/usr/bin/env bash
# Regenerate the Go lexer/parser from the shared DataSetSearch grammar.
#
# The .g4 files in metaquery/search/grammar/ are copies of the single source of
# truth in the sibling `dataset` project (../dataset/src/main/antlr4). Keeping
# one grammar for the Kotlin and Go targets guarantees parse fidelity.
#
# Requires Java + ANTLR 4.13.2. Resolution order:
#   1. $ANTLR_COMPLETE_JAR (a self-contained antlr-4.13.2-complete.jar)
#   2. The tool jar + deps from a local Maven repo ($HOME/.m2)
set -euo pipefail

VERSION=4.13.2
GRAMMAR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../metaquery/search/grammar" && pwd)"
OUT_DIR="$(cd "$GRAMMAR_DIR/.." && pwd)"

run() {
  java -cp "$1" org.antlr.v4.Tool \
    -Dlanguage=Go -package search -o "$OUT_DIR" \
    -lib "$GRAMMAR_DIR" \
    "$GRAMMAR_DIR/DataSetSearchLexer.g4" "$GRAMMAR_DIR/DataSetSearchParser.g4"
}

if [[ -n "${ANTLR_COMPLETE_JAR:-}" ]]; then
  run "$ANTLR_COMPLETE_JAR"
else
  M2="${M2_REPO:-$HOME/.m2/repository}"
  TOOL="$M2/org/antlr/antlr4/$VERSION/antlr4-$VERSION.jar"
  [[ -f "$TOOL" ]] || { echo "ANTLR tool jar not found: $TOOL (set ANTLR_COMPLETE_JAR)" >&2; exit 1; }
  CP="$(find "$M2" \
    -name "antlr4-$VERSION.jar" -o \
    -name "antlr4-runtime-$VERSION.jar" -o \
    -name 'antlr-runtime-3.5.3.jar' -o \
    -name 'ST4-*.jar' -o \
    -ipath '*org.abego.treelayout*' -name '*.jar' -o \
    -ipath '*com/ibm/icu/icu4j*' -name 'icu4j-*.jar' -o \
    -ipath '*javax/json*' -name '*.jar' | tr '\n' ':')"
  run "$CP"
fi

# ANTLR also emits .interp/.tokens (IDE metadata, unused at runtime) — drop them.
rm -f "$OUT_DIR"/*.interp "$OUT_DIR"/*.tokens

echo "Regenerated Go parser in $OUT_DIR"
