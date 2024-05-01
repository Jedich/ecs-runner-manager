package logs

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func NewLogger() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339}

	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger().Level(zerolog.TraceLevel)
}

func Info(msg string) {
	log.Info().Msg(msg)
}

func InfoF(format string, args ...interface{}) {
	log.Info().Msgf(format, args...)
}

func Error(err error) {
	log.Error().Err(err)
}

func ErrorF(format string, args ...interface{}) {
	log.Error().Msgf(format, args...)
}

func Fatal(err error) {
	log.Fatal().Err(err)
}
