package stdlib

import (
	"QIQ/cmd/qiq/runtime"
	"QIQ/cmd/qiq/runtime/stdlib/array"
	"QIQ/cmd/qiq/runtime/stdlib/classes"
	"QIQ/cmd/qiq/runtime/stdlib/ctype"
	"QIQ/cmd/qiq/runtime/stdlib/dateTime"
	"QIQ/cmd/qiq/runtime/stdlib/directory"
	"QIQ/cmd/qiq/runtime/stdlib/errorHandling"
	"QIQ/cmd/qiq/runtime/stdlib/filesystem"
	"QIQ/cmd/qiq/runtime/stdlib/functionHandling"
	"QIQ/cmd/qiq/runtime/stdlib/math"
	"QIQ/cmd/qiq/runtime/stdlib/misc"
	"QIQ/cmd/qiq/runtime/stdlib/optionsInfo"
	"QIQ/cmd/qiq/runtime/stdlib/outputControl"
	"QIQ/cmd/qiq/runtime/stdlib/strings"
	"QIQ/cmd/qiq/runtime/stdlib/variableHandling"
)

func Register(environment runtime.Environment) {
	array.Register(environment)
	classes.Register(environment)
	ctype.Register(environment)
	dateTime.Register(environment)
	directory.Register(environment)
	errorHandling.Register(environment)
	filesystem.Register(environment)
	functionHandling.Register(environment)
	math.Register(environment)
	misc.Register(environment)
	optionsInfo.Register(environment)
	outputControl.Register(environment)
	strings.Register(environment)
	variableHandling.Register(environment)
}
