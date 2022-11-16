package arguments

import (
	"github.com/docopt/docopt.go"
)

var Arguments map[string]any
var VERSION_STRING string

func InitArguments(argv []string) (err error) {

	if Arguments, err = docopt.Parse(commands, argv, false, VERSION_STRING, false, false); err != nil {
		return err
	}

	return
}
