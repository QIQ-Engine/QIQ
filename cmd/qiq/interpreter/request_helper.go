package interpreter

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/config"
	"QIQ/cmd/qiq/ini"
	"QIQ/cmd/qiq/phpError"
	"QIQ/cmd/qiq/request"
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/values"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func parseCookies(cookies string, interpreter runtime.Interpreter) *values.Array {
	result := values.NewArray()

	for cookies != "" {
		if int64(len(result.Keys)) >= interpreter.GetIni().GetInt("max_input_vars") {
			break
		}

		var cookie string
		cookie, cookies, _ = strings.Cut(cookies, ";")
		if cookie == "" {
			continue
		}

		var name string
		var value string
		if !strings.Contains(cookie, "=") {
			// Cookie without value is an empty string
			name = cookie
			value = ""
		} else {
			// Get parameter with key-value-pair
			name, value, _ = strings.Cut(cookie, "=")
		}
		name = strings.Trim(name, " ")
		key := values.NewStr(strings.NewReplacer(
			" ", "_",
			"[", "_",
			".", "_",
		).Replace(name))
		if result.Contains(key) {
			continue
		}
		// Escape plus sign so that it will not be replaced with space
		value = strings.ReplaceAll(value, "+", "%2b")
		value, err := url.QueryUnescape(fixPercentEscaping(value))
		if err != nil {
			if config.IsDevMode {
				println("parseCookies: ", err)
			}
			continue
		}
		result.SetElement(key, values.NewStr(value))
	}

	return result
}

func parsePost(query string, interpreter runtime.Interpreter) (*values.Array, *values.Array, error) {
	postArray := values.NewArray()
	filesArray := values.NewArray()

	query = strings.TrimSuffix(query, "\n")

	if strings.HasPrefix(query, "Content-Type: multipart/form-data;") {
		// TODO Improve code
		var boundary string
		lines := strings.Split(query, "\n")
		lineNum := 0

		contentLen := len(query) - len(lines[lineNum]) - 1
		if contentLen > getPostMaxSize(interpreter.GetIni()) {
			interpreter.PrintError(phpError.NewWarning(
				"PHP Request Startup: POST Content-Length of %d bytes exceeds the limit of %d bytes in Unknown on line 0",
				contentLen,
				getPostMaxSize(interpreter.GetIni()),
			))
			return postArray, filesArray, nil
		}

		for {
			if lineNum >= len(lines) {
				break
			}

			if lineNum == 0 {
				boundary = strings.Replace(lines[lineNum], "Content-Type: multipart/form-data;", "", 1)
				boundary = strings.Replace(strings.TrimSpace(boundary), "boundary=", "", 1)
				if strings.HasPrefix(boundary, `"`) {
					boundary = boundary[1:]
					if strings.Contains(boundary, `"`) {
						boundary = boundary[:strings.Index(boundary, `"`)]
					}
				}
				boundary = "--" + boundary
				if strings.Contains(boundary, ";") {
					// Content-Type: multipart/form-data; boundary=abc; charset=...
					boundary = boundary[:strings.Index(boundary, ";")]
				} else if strings.Contains(boundary, ",") {
					// Content-Type: multipart/form-data; boundary=abc, charset=...
					boundary = boundary[:strings.Index(boundary, ",")]
				}
				lineNum++
				continue
			}

			if lines[lineNum] == boundary+"--" {
				return postArray, filesArray, nil
			}

			if lines[lineNum] == boundary {
				lineNum++
				if strings.HasPrefix(lines[lineNum], "Content-Disposition: form-data") ||
					strings.HasPrefix(lines[lineNum], "Content-Disposition: form-data;") {
					isFile := strings.Contains(lines[lineNum], "filename=")
					fullname := strings.TrimPrefix(lines[lineNum], "Content-Disposition: form-data;")
					fullname = strings.TrimPrefix(fullname, "Content-Disposition: form-data")

					name := strings.Replace(strings.TrimSpace(fullname), "name=", "", 1)
					if strings.Contains(name, ";") {
						name = name[:strings.Index(name, ";")]
					}
					if strings.HasPrefix(name, "'") {
						name = name[1:strings.LastIndex(name, "'")]
						name = strings.ReplaceAll(name, `\'`, "'")
					}
					if strings.HasPrefix(name, `"`) {
						name = name[1:strings.LastIndex(name, `"`)]
						name = strings.ReplaceAll(name, `\"`, `"`)
					}
					name = strings.ReplaceAll(name, `\\`, `\`)
					name = recode(name, interpreter.GetIni())

					filename := ""
					contentType := ""
					if isFile {
						filename = fullname[strings.Index(fullname, "filename="):]
						filename = strings.TrimPrefix(filename, "filename=")
						if strings.HasPrefix(filename, `"`) {
							filename = filename[1:strings.LastIndex(filename, `"`)]
						}
						filename = recode(filename, interpreter.GetIni())
						lineNum++
						if after, ok := strings.CutPrefix(lines[lineNum], "Content-Type:"); ok {
							contentType = after
							contentType = strings.TrimSpace(contentType)
						}
					}

					lineNum += 2
					content := ""
					for lineNum < len(lines) && lines[lineNum] != boundary && lines[lineNum] != boundary+"--" {
						content += lines[lineNum]
						if isFile {
							content += "\n"
						}
						lineNum++
					}
					content = strings.TrimSuffix(content, "\n")
					content = recode(content, interpreter.GetIni())

					if !isFile {
						postArray.SetElement(values.NewStr(name), values.NewStr(content))
						continue
					}
					if isFile && interpreter.GetIni().GetBool("file_uploads") {
						var fileError int64 = 0
						if len(content) > getMaxFilesize(interpreter.GetIni()) {
							content = ""
							fileError = 1
							contentType = ""
						}

						tmpFile := ""
						if fileError == 0 {
							tmpDir := interpreter.GetIni().GetStr("upload_tmp_dir")
							if tmpDir == "" {
								tmpDir = filepath.Join(os.TempDir(), "qiq", "uploads")
							}
							tmpFile = filepath.Join(tmpDir, randomFilename())
							interpreter.GetRequest().UploadedFiles = append(interpreter.GetRequest().UploadedFiles, tmpFile)
							err := common.WriteFile(tmpFile, content)
							if err != nil {
								return postArray, filesArray, fmt.Errorf("parsePost - Failed to write file content: %s", err)
							}
						}

						data := values.NewArray()
						data.SetElement(values.NewStr("name"), values.NewStr(filename))
						data.SetElement(values.NewStr("full_path"), values.NewStr(filename))
						data.SetElement(values.NewStr("type"), values.NewStr(contentType))
						data.SetElement(values.NewStr("tmp_name"), values.NewStr(tmpFile))
						data.SetElement(values.NewStr("error"), values.NewInt(fileError))
						data.SetElement(values.NewStr("size"), values.NewInt(int64(len(content))))

						filesArray.SetElement(values.NewStr(name), data)
					}
					continue
				}
				lineNum++
				continue
			}
			return postArray, filesArray, fmt.Errorf("parsePost - Unexpected line %d: %s", lineNum, lines[lineNum])
		}
		return postArray, filesArray, nil
	}

	postArray, err := parseQuery(strings.ReplaceAll(query, "\n", ""), interpreter)
	return postArray, filesArray, err
}

func randomFilename() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789_-"
	b := make([]byte, 30)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func parseQuery(query string, interpreter runtime.Interpreter) (*values.Array, error) {
	result := values.NewArray()

	for query != "" {
		if int64(len(result.Keys)) >= interpreter.GetIni().GetInt("max_input_vars") {
			break
		}

		var key string
		key, query, _ = strings.Cut(query, interpreter.GetIni().GetStr("arg_separator.input"))
		if key == "" {
			continue
		}

		// Get parameters without key e.g. ab+cd+ef
		// TODO this is only correct if it is in "phpt mode". "Normal" GET will parse it differently
		// ab+cd+ef => array(1) { ["ab_cd_ef"]=> string(0) "" }
		if !strings.Contains(key, "=") && strings.Contains(key, "+") {
			parts := strings.Split(key, "+")
			for i := range parts {
				if len(parts[i]) > getPostMaxSize(interpreter.GetIni()) {
					interpreter.PrintError(phpError.NewWarning(
						"PHP Request Startup: POST Content-Length of %d bytes exceeds the limit of %d bytes in Unknown on line 0",
						len(parts[i]),
						getPostMaxSize(interpreter.GetIni()),
					))
					continue
				}
				if err := result.SetElement(nil, values.NewStr(parts[i])); err != nil {
					return result, err
				}
			}
			continue
		}

		if len(key) > getPostMaxSize(interpreter.GetIni()) {
			interpreter.PrintError(phpError.NewWarning(
				"PHP Request Startup: POST Content-Length of %d bytes exceeds the limit of %d bytes in Unknown on line 0",
				len(key),
				getPostMaxSize(interpreter.GetIni()),
			))
			continue
		}

		// Get parameter with key-value-pair
		key, value, _ := strings.Cut(key, "=")

		key, err := url.QueryUnescape(fixPercentEscaping(key))
		if err != nil {
			return result, err
		}

		value, err = url.QueryUnescape(value)
		if err != nil {
			return result, err
		}
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			result, err = parseQueryKey(key, value, result, interpreter.GetIni())
			if err != nil {
				return result, err
			}
		} else {
			key = strings.NewReplacer(
				" ", "_",
				"+", "_",
				"[", "_",
				".", "_",
			).Replace(key)

			var keyValue values.RuntimeValue
			if common.IsIntegerLiteral(key, false) {
				intValue, _ := common.IntegerLiteralToInt64(key, false)
				keyValue = values.NewInt(intValue)
			} else {
				keyValue = values.NewStr(key)
			}
			result.SetElement(keyValue, values.NewStr(value))
		}
	}

	return result, nil
}

func parseQueryKey(key string, value string, result *values.Array, curIni *ini.Ini) (*values.Array, error) {
	// The parsing of a complex key with arrays is solved by using the interpreter itself:
	// The key and value is transformed into valid PHP code and executed.
	// Example:
	//   Input: 123[][12][de]=abc
	//   Key:   123[][12][de]
	//   Value: abc
	//   PHP:   $array[123][][12]["de"] = "abc";

	firstKey, key, _ := strings.Cut(key, "[")
	key = "[" + key

	maxDepth := curIni.GetInt("max_input_nesting_level")

	phpArrayKeys := []string{firstKey}

	for key != "" {
		key = strings.TrimPrefix(key, "[")
		var nextKey string
		nextKey, key, _ = strings.Cut(key, "]")
		phpArrayKeys = append(phpArrayKeys, nextKey)
		for key != "" && !strings.HasPrefix(key, "[") {
			key = key[1:]
		}
	}

	var php strings.Builder
	php.WriteString("<?php $array")
	for depth, phpArrayKey := range phpArrayKeys {
		if depth+1 >= int(maxDepth) {
			return result, nil
		}
		if phpArrayKey == "" {
			php.WriteString("[]")
		} else if common.IsIntegerLiteral(phpArrayKey, false) {
			phpArrayKeyInt, _ := common.IntegerLiteralToInt64(phpArrayKey, false)
			php.WriteString(fmt.Sprintf("[%d]", phpArrayKeyInt))
		} else {
			php.WriteString("['" + phpArrayKey + "']")
		}
	}
	php.WriteString(" = '" + value + "';")

	interpreter, err := NewInterpreter(runtime.NewExecutionContext(), ini.NewDefaultIni(), &request.Request{}, "")
	if err != nil {
		return nil, err
	}
	interpreter.env.declareVariable("$array", result)
	_, err = interpreter.Process(php.String())

	return interpreter.env.variables["$array"].Value.(*values.Array), err
}

// This fix is required because "url.QueryUnescape()" cannot handle an unescaped percent
func fixPercentEscaping(key string) string {
	re, _ := regexp.Compile("%([^0-9A-Fa-f]|$)")
	// Replace only the '%' character with '%25' without affecting the following character
	return re.ReplaceAllStringFunc(key, func(match string) string {
		return "%25" + match[1:]
	})
}

func getMaxFilesize(ini *ini.Ini) int {
	sizeStr := ini.GetStr("upload_max_filesize")
	if common.IsDecimalLiteral(sizeStr, false) {
		size := int(common.DecimalLiteralToInt64(sizeStr, false))
		if size == 0 {
			return 2 * 1024 * 1024
		}
		return size
	}
	return resolveSiPrefix(sizeStr, 2*1024*1024)
}

func getPostMaxSize(ini *ini.Ini) int {
	sizeStr := ini.GetStr("post_max_size")
	if common.IsDecimalLiteral(sizeStr, false) {
		size := int(common.DecimalLiteralToInt64(sizeStr, false))
		if size == 0 {
			return 8 * 1024 * 1024
		}
		return size
	}
	return resolveSiPrefix(sizeStr, 8*1024*1024)
}

func resolveSiPrefix(number string, defaultResult int) int {
	if strings.HasSuffix(number, "K") {
		return int(common.DecimalLiteralToInt64(strings.Replace(number, "K", "", 1), false) * 1024)
	}
	if strings.HasSuffix(number, "M") {
		return int(common.DecimalLiteralToInt64(strings.Replace(number, "M", "", 1), false) * 1024 * 1024)
	}
	if strings.HasSuffix(number, "G") {
		return int(common.DecimalLiteralToInt64(strings.Replace(number, "G", "", 1), false) * 1024 * 1024 * 1024)
	}
	return defaultResult
}

func recode(input string, ini *ini.Ini) string {
	if !ini.GetBool("mbstring.encoding_translation") {
		return input
	}

	var decoder *transform.Reader
	var encoder transform.Transformer

	inputEncoding := ini.GetStr("input_encoding")
	if inputEncoding == "" {
		inputEncoding = ini.GetStr("default_charset")
	}
	switch inputEncoding {
	case "Shift_JIS":
		decoder = transform.NewReader(strings.NewReader(input), japanese.ShiftJIS.NewDecoder())
	case "UTF-8":
		decoder = transform.NewReader(strings.NewReader(input), unicode.UTF8.NewDecoder())
	default:
		fmt.Println("changeEncoding: Unsupported from encoding: ", inputEncoding)
		return ""
	}

	decodedBytes, err := io.ReadAll(decoder)
	if err != nil {
		fmt.Println("changeEncoding: Error decoding input: ", err)
		return ""
	}

	internalEncoding := ini.GetStr("internal_encoding")
	if internalEncoding == "" {
		internalEncoding = ini.GetStr("default_charset")
	}
	switch internalEncoding {
	case "Shift_JIS":
		encoder = japanese.ShiftJIS.NewEncoder()
	case "UTF-8":
		encoder = unicode.UTF8.NewEncoder()
	default:
		fmt.Println("changeEncoding: Unsupported from encoding: ", internalEncoding)
		return ""
	}

	encodedBytes, err := io.ReadAll(transform.NewReader(strings.NewReader(string(decodedBytes)), encoder))
	if err != nil {
		fmt.Println("changeEncoding: Error encoding output: ", err)
		return ""
	}

	return string(encodedBytes)
}
