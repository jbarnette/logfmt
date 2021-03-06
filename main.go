package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
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

// simple strings are not quoted
var simple = regexp.MustCompile("^[-+_/:|@.a-zA-Z0-9]+$")

func main() {
	var excluded globs

	flag.Var(&excluded, "x",
		"Don't print keys matching this glob")

	flag.Parse()

	color := isatty.IsTerminal(os.Stdout.Fd())
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
			if color {
				if _, err := w.Write([]byte("\033[0;33m")); err != nil {
					panic(err)
				}
			}

			if err := writeLine(os.Stderr, line); err != nil {
				panic(err)
			}

			if color {
				if _, err := w.Write([]byte("\033[0m")); err != nil {
					panic(err)
				}
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

			if err := writeKeyValue(w, k, data[k], color); err != nil {
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

// sortedKeys returns a slice containing data's keys, sorted mostly alphabetically.
func sortedKeys(data map[string]interface{}) (keys []string) {
	for k := range data {
		if k != "at" {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	if _, ok := data["at"]; ok {
		keys = append([]string{"at"}, keys...)
	}

	return
}

func writeLine(w io.Writer, line []byte) error {
	if _, err := w.Write(line); err != nil {
		return err
	}

	_, err := w.Write([]byte("\n"))
	return err
}

// writeKeyValue writes "key=value" to w. The "key=" part is highlighted if color is true.
func writeKeyValue(w io.Writer, key string, value interface{}, color bool) error {
	if color {
		if _, err := w.Write([]byte("\033[0;36m")); err != nil {
			return err
		}
	}

	if err := writeString(w, key); err != nil {
		return err
	}

	if _, err := w.Write([]byte("=")); err != nil {
		return err
	}

	if color {
		if _, err := w.Write([]byte("\033[0m")); err != nil {
			return err
		}
	}

	if err := writeValue(w, value); err != nil {
		return err
	}

	return nil
}

func writeValue(w io.Writer, value interface{}) error {
	switch v := value.(type) {
	case bool:
		return writeBool(w, v)

	case float64:
		return writeFloat64(w, v)

	case string:
		return writeString(w, v)

	case nil:
		return writeNull(w)

	case []interface{}:
		return writeArray(w, v)

	case map[string]interface{}:
		return writeObject(w, v)

	default:
		return fmt.Errorf("logfmt: unsupported value type: %T", v)
	}
}

func writeBool(w io.Writer, b bool) error {
	_, err := w.Write([]byte(strconv.FormatBool(b)))
	return err
}

func writeFloat64(w io.Writer, f float64) error {
	_, err := w.Write([]byte(strconv.FormatFloat(f, 'g', 4, 64)))
	return err
}

func writeString(w io.Writer, s string) error {
	if !simple.MatchString(s) {
		s = strconv.Quote(s)
	}
	_, err := w.Write([]byte(s))
	return err
}

func writeNull(w io.Writer) error {
	_, err := w.Write([]byte("null"))
	return err
}

func writeArray(w io.Writer, a []interface{}) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}

	for i, v := range a {
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

func writeObject(w io.Writer, o map[string]interface{}) error {
	if _, err := w.Write([]byte("{")); err != nil {
		return err
	}

	for i, k := range sortedKeys(o) {
		if i > 0 {
			if _, err := w.Write([]byte(" ")); err != nil {
				return err
			}
		}

		if err := writeKeyValue(w, k, o[k], false); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("}")); err != nil {
		return err
	}

	return nil
}
