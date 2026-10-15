## What this changes

<!-- One or two sentences. Link the issue if there is one. -->

## Why

<!-- What was wrong, or what this makes possible. -->

## Checklist

- [ ] `gofmt -l .` prints nothing
- [ ] `go vet ./...` passes
- [ ] `go test -race ./...` passes
- [ ] Added or updated a test that fails without this change
- [ ] Updated the README if a flag, an output format, or the parser's behaviour changed
- [ ] Regenerated the README example tree if the repository layout changed
- [ ] `treegen | treegen new` still round-trips, if the walk or the renderer changed

<!-- There is no CI. Please actually run the above. -->
