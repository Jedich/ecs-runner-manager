package logs

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func NewLogger() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.TraceLevel)
}

func Info(msg string) {
	log.Info().Msg(msg)
}

func InfoF(format string, args ...interface{}) {
	log.Info().Msgf(format, args...)
}

func Error(err error) {
	log.Error().Err(err).Send()
}

func ErrorF(format string, args ...interface{}) {
	log.Error().Msgf(format, args...)
}

func Fatal(err error) {
	log.Fatal().Err(err).Send()
}
