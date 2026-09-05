# Runs `go test -race` inside a Linux container (the Windows race detector
# needs a C toolchain; Docker sidesteps that). Requires Docker Desktop running.
#
# Most of this track is single-goroutine, so `check.ps1` is the usual gate.
# Reach for this one when a module or exercise touches shared memory across
# goroutines - Module 1's map-write test and Module 3's method-set work do.
#
# Usage:
#   .\test-race.ps1                            # whole repo
#   .\test-race.ps1 ./01-memory/...            # one module
#   .\test-race.ps1 -run TestRing ./01-memory/exercises
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$TestArgs
)

if (-not $TestArgs) { $TestArgs = @('./...') }

docker run --rm `
    -v "${PSScriptRoot}:/app" `
    -w /app `
    -v go-mastery-build-cache:/root/.cache/go-build `
    -v go-mastery-mod-cache:/go/pkg/mod `
    golang:1.26 `
    go test -race @TestArgs

exit $LASTEXITCODE
