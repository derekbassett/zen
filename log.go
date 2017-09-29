package zen

import (
	"io"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

// SetLogLevel support
// panic
// fatal
// error
// warn, warning
// info
// debug
func SetLogLevel(level string) {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	log.SetLevel(lvl)
}

// SetLogOutput to an io.Writer
func SetLogOutput(out io.Writer) {
	log.SetOutput(out)
}
