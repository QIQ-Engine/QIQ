package ini

import (
	"QIQ/cmd/qiq/common"
	"QIQ/cmd/qiq/phpError"
	"maps"
	"slices"
	"strings"
)

// Spec: https://www.php.net/manual/en/info.constants.php#constant.ini-system
const (
	// Entry can be set in user scripts (like with ini_set()) or in the Windows registry. Entry can be set in .user.ini
	INI_USER int = 1
	// Entry can be set in php.ini, .htaccess, httpd.conf or .user.ini
	INI_PERDIR int = 2
	// Entry can be set in php.ini or httpd.conf
	INI_SYSTEM int = 4
	// Entry can be set anywhere
	INI_ALL int = 7
)

type Ini struct {
	directives map[string]string
}

func NewDefaultIni() *Ini { return &Ini{directives: copyDefaultValues()} }

func NewDevIni() *Ini {
	defaultIni := NewDefaultIni()
	defaultIni.SetToDev()
	return defaultIni
}

func NewDevIniFromArray(ini []string) (*Ini, phpError.Error) {
	defaultIni := NewDevIni()
	var resultErr phpError.Error = nil

	for _, setting := range ini {
		parts := strings.SplitN(setting, "=", 2)
		err := defaultIni.Set(parts[0], parts[1], INI_ALL)
		if err != nil {
			if resultErr == nil {
				resultErr = phpError.NewError("%s", err)
			} else {
				resultErr = phpError.NewError("%s\n%s", resultErr, err)
			}
		}
	}

	return defaultIni, resultErr
}

func (ini *Ini) SetToDev() {
	ini.Set("error_reporting", "32767", INI_ALL)
	ini.Set("expose_php", "1", INI_ALL)
}

func (ini *Ini) Set(directive string, value string, source int) phpError.Error {
	changeable, found := allowedDirectives[directive]
	if !found {
		return phpError.NewError("Directive %s not found", directive)
	}

	if changeable&source == 0 {
		return phpError.NewError("Not allowed to change %s", directive)
	}

	if IsBool(directive) {
		if value == "1" || strings.ToLower(value) == "on" {
			ini.directives[directive] = "1"
			return nil
		}
		ini.directives[directive] = ""
		return nil
	}

	if slices.Contains(intDirectives, directive) {
		if !common.IsIntegerLiteralWithSign(value, false) {
			return nil
		}
		ini.directives[directive] = value
		return nil
	}

	ini.directives[directive] = value
	return nil
}

func (ini *Ini) Get(directive string) (string, phpError.Error) {
	if _, found := allowedDirectives[directive]; !found {
		return "", phpError.NewError("Directive not found")
	}

	return ini.directives[directive], nil
}

func IsBool(directive string) bool {
	return slices.Contains(boolDirectives, directive)
}

func (ini *Ini) GetBool(directive string) bool {
	value, err := ini.Get(directive)
	if err != nil {
		return false
	}
	return value == "1"
}

func (ini *Ini) GetInt(directive string) int64 {
	value, err := ini.Get(directive)
	if err != nil {
		return -1
	}
	if common.IsIntegerLiteralWithSign(value, false) {
		intVal, _ := common.IntegerLiteralToInt64WithSign(value, false)
		return intVal
	}
	return -1
}

func (ini *Ini) GetStr(directive string) string {
	value, err := ini.Get(directive)
	if err != nil {
		return ""
	}
	return value
}

func GetDirectives() []string {
	result := []string{}
	for directive := range allowedDirectives {
		result = append(result, directive)
	}
	return result
}

func copyDefaultValues() map[string]string {
	copyMap := make(map[string]string, len(defaultValues))
	maps.Copy(copyMap, defaultValues)
	return copyMap
}
