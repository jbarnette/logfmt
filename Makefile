.PHONY: test

cmd = go run . -x ignored.*

test:
	$(cmd) < test.jsonl 2>&1 | diff -u test.out -

test.out: test.jsonl
	$(cmd) < $< 2>&1 | cat > $@
