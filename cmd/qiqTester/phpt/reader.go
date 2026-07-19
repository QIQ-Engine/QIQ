package phpt

import (
	"QIQ/cmd/qiq/common"
	replacejson "QIQ/cmd/qiqTester/replaceJson"
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Reader struct {
	filename    string
	replaceJson replacejson.ReplaceEntry
	sections    []string
	lines       []string
	currentLine int
	testFile    *TestFile
}

func NewReader(filename string, replaceJson replacejson.ReplaceEntry) (*Reader, error) {
	lines := []string{}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return &Reader{filename: filename, replaceJson: replaceJson, sections: []string{}, lines: lines, currentLine: 0, testFile: NewTestFile(filename)}, nil
}

func (reader *Reader) GetTestFile() (*TestFile, error) {
	for !reader.isEof() {
		if reader.at() == "--TEST--" {
			reader.eat()
			// According to spec the title is a single line
			// Spec: https://qa.php.net/phpt_details.php#test_section

			// But the official PHP test runner can also handle multi line titles
			// See: https://github.com/php/php-src/issues/17761
			reader.testFile.Title = reader.eat()
			for !reader.isEof() && !reader.isSection(reader.at()) {
				reader.testFile.Title += reader.eat()
			}
			reader.sections = append(reader.sections, "--TEST--")
			continue
		}

		if slices.Contains([]string{"--CREDITS--", "--EXTENSIONS--", "--DESCRIPTION--"}, reader.at()) {
			reader.eat()
			for !reader.isEof() && !reader.isSection(reader.at()) {
				reader.eat()
			}
			continue
		}

		if reader.at() == "--SKIPIF--" {
			reader.eat()
			file := ""
			for !reader.isEof() && !reader.isSection(reader.at()) {
				if file != "" {
					file += "\n"
				}
				file += reader.eat()
			}
			reader.testFile.File = file
			reader.sections = append(reader.sections, "--SKIPIF--")
			continue
		}

		if reader.at() == "--POST--" || reader.at() == "--POST_RAW--" {
			section := reader.eat()
			var postData strings.Builder
			for !reader.isEof() && !reader.isSection(reader.at()) {
				postData.WriteString(reader.eat() + "\n")
			}
			reader.testFile.Post = postData.String()
			reader.sections = append(reader.sections, section)
			continue
		}

		if reader.at() == "--GET--" {
			reader.eat()
			reader.testFile.Get = reader.eat()
			reader.sections = append(reader.sections, "--GET--")
			continue
		}

		if reader.at() == "--COOKIE--" {
			reader.eat()
			reader.testFile.Cookie = reader.eat()
			reader.sections = append(reader.sections, "--COOKIE--")
		}

		if reader.at() == "--INI--" {
			ini := []string{}
			reader.eat()
			for !reader.isEof() && !reader.isSection(reader.at()) {
				ini = append(ini, reader.eat())
			}
			reader.testFile.Ini = ini
			reader.sections = append(reader.sections, "--INI--")
			continue
		}

		if reader.at() == "--ARGS--" {
			reader.eat()
			argsStr := reader.eat()
			parts := strings.Split(argsStr, " ")
			args := [][]string{{reader.filename}}
			for i := range parts {
				args = append(args, []string{parts[i]})
			}
			// TODO parse --arg value --arg=value -avalue -a=value -a value
			reader.testFile.Args = args
			reader.sections = append(reader.sections, "--ARGS--")
			continue
		}

		if reader.at() == "--ENV--" {
			reader.eat()
			env := map[string]string{}
			for !reader.isEof() && !reader.isSection(reader.at()) {
				line := reader.eat()
				separator := strings.Index(line, "=")
				env[line[:separator]] = line[separator+1:]
			}
			reader.testFile.Env = env
			reader.sections = append(reader.sections, "--ENV--")
			continue
		}

		if reader.at() == "--FILE--" {
			reader.eat()
			var file strings.Builder
			file.WriteString(reader.testFile.File)
			for !reader.isEof() && !reader.isSection(reader.at()) {
				file.WriteString(reader.eat() + "\n")
			}
			reader.testFile.File = file.String()
			reader.sections = append(reader.sections, "--FILE--")
			continue
		}

		if reader.at() == "--EXPECT--" || reader.at() == "--EXPECTF--" || reader.at() == "--EXPECTREGEX--" {
			section := reader.eat()
			expect := ""
			for !reader.isEof() && !reader.isSection(reader.at()) {
				expect += reader.eat() + "\n"
			}

			// Replacements
			if reader.replaceJson.Section == section {
				expect = strings.ReplaceAll(expect, reader.replaceJson.Search, reader.replaceJson.Replace)
			}

			reader.testFile.Expect = common.TrimTrailingLineBreaks(expect)
			reader.testFile.ExpectType = section
			reader.sections = append(reader.sections, section)
			continue
		}

		if reader.at() == "--CLEAN--" {
			reader.eat()
			var file strings.Builder
			file.WriteString(reader.testFile.File)
			for !reader.isEof() && !reader.isSection(reader.at()) {
				file.WriteString(reader.eat() + "\n")
			}
			reader.testFile.File = file.String()
			continue
		}

		// Spec: https://php.github.io/php-src/miscellaneous/writing-tests.html#cgi
		// This section takes no value. It merely provides a simple marker for tests
		// that MUST be run as CGI, even if there is no --POST-- or --GET-- sections in the test file.
		if reader.at() == "--CGI--" {
			reader.eat()

			// TODO qiqTester - how to handle CGI?
			continue
		}

		return reader.testFile, fmt.Errorf(`Unsupported section "%s"`, reader.at())
	}

	if err := reader.isValid(); err != nil {
		return nil, err
	}

	return reader.testFile, nil
}

func (reader *Reader) isValid() error {
	if !reader.hasSection("--TEST--") {
		return fmt.Errorf(`Section "--TEST--" is missing`)
	}

	if !reader.hasSection("--FILE--") && !reader.hasSection("--FILEEOF--") &&
		!reader.hasSection("--FILE_EXTERNAL--") && !reader.hasSection("--REDIRECTTEST--") {
		return fmt.Errorf(`Section "--FILE-- | --FILEEOF-- | --FILE_EXTERNAL-- | --REDIRECTTEST--" is missing`)
	}

	if !reader.hasSection("--EXPECT--") && !reader.hasSection("--EXPECTF--") && !reader.hasSection("--EXPECTREGEX--") &&
		!reader.hasSection("--EXPECT_EXTERNAL--") && !reader.hasSection("--EXPECTF_EXTERNAL--") &&
		!reader.hasSection("--EXPECTREGEX_EXTERNAL--") {
		return fmt.Errorf(`Section "--EXPECT-- | --EXPECTF-- | --EXPECTREGEX-- | --EXPECT_EXTERNAL-- | --EXPECTF_EXTERNAL-- | --EXPECTREGEX_EXTERNAL--" is missing`)
	}
	return nil
}

func (reader *Reader) hasSection(section string) bool {
	return slices.Contains(reader.sections, section)
}

var sections = []string{
	"--TEST--", "--DESCRIPTION--", "--CREDITS--", "--SKIPIF--", "--CONFLICTS--", "--WHITESPACE_SENSITIVE--", "--CAPTURE_STDIO--",
	"--EXTENSIONS--", "--POST--", "--PUT--", "--POST_RAW--", "--GZIP_POST--", "--DEFLATE_POST--", "--GET--", "--COOKIE--",
	"--STDIN--", "--INI--", "--ARGS--", "--ENV--", "--PHPDBG--", "--FILE--", "--FILEEOF--", "--FILE_EXTERNAL--",
	"--REDIRECTTEST--", "--CGI--", "--XFAIL--", "--EXPECTHEADERS--", "--EXPECT--", "--EXPECTF--", "--EXPECTREGEX--",
	"--EXPECT_EXTERNAL--", "--EXPECTF_EXTERNAL--", "--EXPECTREGEX_EXTERNAL--", "--CLEAN--",
}

func (reader *Reader) isSection(section string) bool {
	return slices.Contains(sections, section)
}

/*
Spec: https://qa.php.net/phpt_details.php
[] indicates optional sections.

--TEST--
[--DESCRIPTION--]
[--CREDITS--]
[--SKIPIF--]
[--CONFLICTS--]
[--WHITESPACE_SENSITIVE--]
[--CAPTURE_STDIO--]
[--EXTENSIONS--]
[--POST-- | --PUT-- | --POST_RAW-- | --GZIP_POST-- | --DEFLATE_POST-- | --GET--]
[--COOKIE--]
[--STDIN--]
[--INI--]
[--ARGS--]
[--ENV--]
[--PHPDBG--]
--FILE-- | --FILEEOF-- | --FILE_EXTERNAL-- | --REDIRECTTEST--
[--CGI--]
[--XFAIL--]
[--EXPECTHEADERS--]
--EXPECT-- | --EXPECTF-- | --EXPECTREGEX-- | --EXPECT_EXTERNAL-- | --EXPECTF_EXTERNAL-- | --EXPECTREGEX_EXTERNAL--
[--CLEAN--]
*/
