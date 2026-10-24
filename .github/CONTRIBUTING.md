# Contributing

Thanks for helping improve **gollback**. By contributing, you agree to abide by the [code of conduct](CODE_OF_CONDUCT.md).

## Issues and questions

Search the [issue tracker](https://github.com/vardius/gollback/issues) first; if nothing matches, open an issue.
Bugs, ideas and usage questions are all welcome. For a bug, include the functions you passed and what you expected back.

## Development

You need Go 1.26 or newer.

```sh
git clone https://github.com/vardius/gollback.git
cd gollback
go test ./...
```

Before sending a pull request, run what CI runs:

```sh
gofmt -l .
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go test -race -shuffle=on ./...
go test -run Allocations -count 1 ./...
```

### Layout

- `gollback.go` holds `Race`, `All` and `Retry`; `doc.go` is the package documentation.
- Tests run on a fake clock with `testing/synctest`, so timings are exact and nothing sleeps for real. Keep new tests that way.
- `examples/` holds runnable programs, each with a test. The website shows them, so keep them short.
- `website/` is the documentation site.

### Performance

`TestAllocations` fails if a call allocates more than it should; if your change moves a number, explain why in the pull request.
To compare before and after:

```sh
go test -run '^$' -bench . -count 10 > new.txt
```

and compare with the same run on `master`, for example with `golang.org/x/perf/cmd/benchstat`.

### Website

```sh
cd website
npm ci
npm run dev
```

## Pull requests

- For a large change, open an issue first so we can agree on the approach.
- Branch from `master`, keep the change focused, and add a test that fails without it.
