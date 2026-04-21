#!/usr/bin/env bash
#
# Publish a tagged release:
#   - builds the WASM blob
#   - computes sha256
#   - creates (or updates) the GitHub release with the .wasm asset
#   - rewrites the release block in README.md between the
#     `<!-- release:begin -->` / `<!-- release:end -->` markers
#
# Invoked by `make release VERSION=vX.Y.Z`. Requires a clean working tree, the
# tag to exist locally, and `gh` CLI authenticated against the repo.
#
# Does NOT auto-commit the README update — caller reviews + commits.

set -euo pipefail

VERSION="${1:-}"
REPO="${REPO:-IodeSystems/sqlc-go-codegen-metaquery}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

die() { echo "release: $*" >&2; exit 1; }
info() { echo "release: $*"; }

if [ -z "$VERSION" ]; then
    die "Usage: scripts/release.sh <version>  (e.g. v0.2.0)"
fi
[[ "$VERSION" == v* ]] || die "version must start with 'v' (got: $VERSION)"

cd "$ROOT"

# Preflight
[ -z "$(git status --porcelain)" ] || die "working tree is dirty — commit or stash first"
gh auth status >/dev/null 2>&1 || die "gh CLI not authenticated — run 'gh auth login'"
git rev-parse "$VERSION" >/dev/null 2>&1 || die "tag $VERSION does not exist locally — create with: git tag -a $VERSION -m 'release $VERSION'"

# Push tag if not yet on origin
if ! git ls-remote --tags origin | grep -q "refs/tags/$VERSION\$"; then
    info "pushing tag $VERSION to origin"
    git push origin "$VERSION"
fi

# Build fresh wasm
info "building WASM"
rm -f bin/sqlc-go-codegen-metaquery.wasm
make bin/sqlc-go-codegen-metaquery.wasm >/dev/null

WASM="bin/sqlc-go-codegen-metaquery.wasm"
SHA_FILE="$WASM.sha256"

sha256sum "$WASM" > "$SHA_FILE"
SHA="$(awk '{print $1}' "$SHA_FILE")"
URL="https://github.com/$REPO/releases/download/$VERSION/sqlc-go-codegen-metaquery.wasm"

info "sha256: $SHA"
info "url:    $URL"

# Create or update release
if gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1; then
    info "release $VERSION exists — uploading assets (clobber)"
    gh release upload "$VERSION" "$WASM" "$SHA_FILE" --clobber --repo "$REPO"
else
    info "creating release $VERSION"
    gh release create "$VERSION" "$WASM" "$SHA_FILE" \
        --repo "$REPO" \
        --title "$VERSION" \
        --notes "Built from tag $VERSION. See README for usage."
fi

# Rewrite README release block
README="$ROOT/README.md"
[ -f "$README" ] || die "README.md not found at $README"
grep -q '<!-- release:begin -->' "$README" || die "README.md missing release markers — expected <!-- release:begin --> / <!-- release:end -->"

awk -v ver="$VERSION" -v url="$URL" -v sha="$SHA" '
/<!-- release:begin -->/ {
    print
    print ""
    printf "### Latest release — %s\n\n", ver
    print "```yaml"
    print "plugins:"
    print "- name: metaquery"
    print "  wasm:"
    printf "    url: %s\n", url
    printf "    sha256: %s\n", sha
    print "```"
    print ""
    in_block = 1
    next
}
/<!-- release:end -->/ {
    in_block = 0
    print
    next
}
in_block { next }
{ print }
' "$README" > "$README.new"
mv "$README.new" "$README"

info "done. README updated."
echo
echo "Next:"
echo "  git add README.md"
echo "  git commit -m 'docs: release $VERSION'"
echo "  git push"
