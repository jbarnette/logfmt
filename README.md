# JSONL → logfmt

This command-line tool reads [JSON lines](https://jsonlines.org) from stdin and prints something like [logfmt](https://brandur.org/logfmt) to stdout. Keys are highlighted if stdout is a terminal. Non-JSON lines are printed to stderr. Pass `-x <glob>` to ignore keys.

```
$ echo '{"at":"hello"}' | go run github.com/jbarnette/logfmt
at=hello
```
