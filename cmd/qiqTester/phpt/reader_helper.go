package phpt

import (
	"fmt"
	"net/url"
	"strings"
)

func (reader *Reader) isEof() bool { return reader.currentLine > len(reader.lines)-1 }

func (reader *Reader) at() string {
	if reader.isEof() {
		return ""
	}
	return reader.lines[reader.currentLine]
}

func (reader *Reader) eat() string {
	if reader.isEof() {
		return ""
	}

	result := reader.at()
	reader.currentLine++
	return result
}

func parseQuery(query string) ([][]string, error) {
	result := [][]string{}

	var err error
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		// Get parameters without key e.g. ab+cd+ef
		if !strings.Contains(key, "=") && strings.Contains(key, "+") {
			parts := strings.Split(key, "+")
			for i := range parts {
				result = append(result, []string{parts[i]})
			}
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}

		result = append(result, []string{key, value})
	}
	return result, err
}
