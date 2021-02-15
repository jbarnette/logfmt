.PHONY: test

test:
	go run . < test.jsonl 2>&1 | diff -u test.out -

test.out: test.jsonl
	go run . < $< 2>&1 | cat > $@
