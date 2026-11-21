# Test plan: st_1m7w06gkd6nha79mt02s6p

## D9 Attestation: `go mod download` as the setup command

RFX x_1m7vzw587djxz0tw207qre decided `go mod download` is the setup command to declare. This block runs the command three times (normal, idempotent, offline) and then builds, confirming a fresh checkout is made ready by one run.

```sh produces=out.log inputs=go.mod,go.sum,setup_test.go
set -eu
{
echo "=== D9 Attestation: go mod download ==="
echo ""
echo "--- Run 1: go mod download (populates module cache) ---"
go mod download
echo "exit: $?"
echo ""
echo "--- Run 2: go mod download again (idempotency) ---"
go mod download
echo "exit: $?"
echo ""
echo "--- Run 3: go mod download with GOPROXY=off (offline) ---"
GOPROXY=off go mod download
echo "exit: $?"
echo ""
echo "--- Build after setup ---"
go build ./...
echo "exit: $?"
} > "$OUT/out.log" 2>&1
cat "$OUT/out.log"
```
