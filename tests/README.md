# Tests

## Unit tests
```
cd tests && go test ./unit/...
```

## Structural tests
```
cd tests && go test ./structural/...
```

## All offline tests
```
cd tests && go test ./...
```

## Behavioral evals (live Codex CLI, via HarnessBench)
```
cd installer && make evals
```

Requires the `codex` CLI on PATH and already authenticated — see `.smaqit/bench/README.md` for the manifest layout and `docs/wiki/benchmarking.md` for the Bench engine itself. Run evidence is written under `.smaqit/bench/runs/`.

This replaces the former Copilot-SDK-based `tests/evals/` runner and JSON corpus (removed in Task 026 after verifying the HarnessBench suite above passes against a live, authenticated Codex CLI — see `.smaqit/bench/MIGRATION.md`).
