package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type globs []string

func (g *globs) String() string {
	return strings.Join(*g, ", ")
}

func (g *globs) Set(value string) error {
	*g = append(*g, value)
	return nil
}

// Match returns true if any of the globs match the provided string.
func (g globs) Match(s string) bool {
	for _, glob := range g {
		if matched, _ := filepath.Match(glob, s); matched {
			return true
		}
	}

	return false
}

func main() {
	var excluded globs

	flag.Var(&excluded, "x",
		"Don't print keys matching this glob")

	flag.Parse()

	r := bufio.NewReaderSize(os.Stdin, 64*1024)
	w := os.Stdout

	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		if len(line) == 0 {
			continue
		}

		if isPrefix {
			panic("logfmt: line > 64K")
		}

		data := map[string]interface{}{}
		if err := json.Unmarshal(line, &data); err != nil {
			if err := writeLine(os.Stderr, line); err != nil {
				panic(err)
			}

			continue
		}

		keys := sortedKeys(data)
		for i, k := range keys {
			if excluded.Match(k) {
				continue
			}

			if i > 0 {
				if _, err := w.Write([]byte(" ")); err != nil {
					panic(err)
				}
			}

			if err := writeKeyValue(w, k, data[k]); err != nil {
				panic(err)
			}
		}

		if len(keys) > 0 {
			if _, err := w.Write([]byte("\n")); err != nil {
				panic(err)
			}
		}
	}
}

// sortedKeys returns a slice containing data's keys, sorted alphabetically.
func sortedKeys(data map[string]interface{}) (keys []string) {
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return
}

func writeLine(w io.Writer, line []byte) error {
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}

func writeKeyValue(w io.Writer, key string, value interface{}) error {
	if err := writeKey(w, key); err != nil {
		return err
	}

	if _, err := w.Write([]byte("=")); err != nil {
		return err
	}

	if err := writeValue(w, value); err != nil {
		return err
	}

	return nil
}

func writeKey(w io.Writer, key string) error {
	if _, err := w.Write([]byte(key)); err != nil {
		return err
	}

	return nil
}

func writeValue(w io.Writer, value interface{}) error {
	switch v := value.(type) {
	case bool:
		return writeBoolValue(w, v)

	case float64:
		return writeFloat64Value(w, v)

	case string:
		return writeStringValue(w, v)

	case nil:
		return writeNullValue(w)

	case []interface{}:
		return writeArrayValue(w, v)

	case map[string]interface{}:
		return writeObjectValue(w, v)

	default:
		return fmt.Errorf("logfmt: unsupported value type: %T", v)
	}
}

func writeBoolValue(w io.Writer, value bool) error {
	_, err := fmt.Fprintf(w, "%v", value)
	return err
}

func writeFloat64Value(w io.Writer, value float64) error {
	_, err := fmt.Fprintf(w, "%v", value)
	return err
}

func writeStringValue(w io.Writer, value string) error {
	_, err := fmt.Fprintf(w, "%v", value)
	return err
}

func writeNullValue(w io.Writer) error {
	_, err := w.Write([]byte("null"))
	return err
}

func writeArrayValue(w io.Writer, values []interface{}) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}

	for i, v := range values {
		if i > 0 {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}

		if err := writeValue(w, v); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("]")); err != nil {
		return err
	}

	return nil
}

func writeObjectValue(w io.Writer, obj map[string]interface{}) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}

	for i, k := range sortedKeys(obj) {
		if i > 0 {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}

		if err := writeKeyValue(w, k, obj[k]); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("]")); err != nil {
		return err
	}

	return nil
}
