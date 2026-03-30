# Tests

## Unit tests
```
cd tests && go test ./unit/...
```

## Structural tests
```
cd tests && go test ./structural/...
```

## All non-eval tests
```
cd tests && go test ./...
```

## Behavioral evals (live Copilot API)
```
cd tests && go run ./evals/runner/... ./evals/
```

Run reports are written to `tests/evals/runs/<YYYYMMDD-HHMMSS>/`.
