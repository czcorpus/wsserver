package actions

import (
	"net/http"

	"github.com/czcorpus/wsserver/core"
	"github.com/rs/zerolog/log"
)

func mapError(err core.AppError) int {
	switch err.Type {
	case core.ErrorTypeInternalError:
		return http.StatusInternalServerError
	case core.ErrorTypeNotFound:
		return http.StatusNotFound
	case core.ErrorTypeInvalidArguments:
		return http.StatusBadRequest
	default:
		log.Warn().Str("errType", string(err.Type)).Msg("encountered an unknown error type")
		return http.StatusInternalServerError
	}

}
