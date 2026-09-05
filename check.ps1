# THE GATE for this track. Four stages, in order, stopping at the first failure:
#
#   1. go build   - it compiles
#   2. go vet     - the standard library's own bug detector
#   3. staticcheck- the deeper static analyser (unused results, wrong loops,
#                   dead code, misuse of stdlib APIs)
#   4. go test    - the behavioural + allocation-budget tests
#
# `go test` alone is NOT the definition of done on this track. Plenty of the
# bugs Module 1 teaches (aliasing, retained backing arrays, per-call
# allocations) compile clean and pass a happy-path test. The analysers and the
# allocation budgets in the tests are what actually catch them.
#
# Usage:
#   .\check.ps1                       # whole repo
#   .\check.ps1 ./01-memory/...       # one module
#   .\check.ps1 -Bench ./01-memory/...# also run benchmarks with alloc counts
param(
    [switch]$Bench,
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Target
)

if (-not $Target) { $Target = @('./...') }

function Invoke-Stage {
    param([string]$Name, [scriptblock]$Body)
    Write-Host ""
    Write-Host "=== $Name ===" -ForegroundColor Cyan
    & $Body
    if ($LASTEXITCODE -ne 0) {
        Write-Host "FAILED at stage: $Name" -ForegroundColor Red
        exit $LASTEXITCODE
    }
}

Invoke-Stage 'go build'     { go build $Target }
Invoke-Stage 'go vet'       { go vet $Target }
Invoke-Stage 'staticcheck'  { staticcheck $Target }
Invoke-Stage 'go test'      { go test $Target }

if ($Bench) {
    Invoke-Stage 'go test -bench' { go test -run '^$' -bench . -benchmem $Target }
}

Write-Host ""
Write-Host "ALL STAGES PASSED" -ForegroundColor Green
