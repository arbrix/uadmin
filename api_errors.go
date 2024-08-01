package uadmin

import (
	"fmt"
	"net/http"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json; charset=utf-8"
	errStatus         = "error"
	defaultErrorMsg   = "unknown server error"
)

var codeToGenericMessage = map[int]string{
	http.StatusBadRequest:          "400. Bad request",
	http.StatusNotFound:            "404. Not found",
	http.StatusInternalServerError: "500. Internal Server Error",
}

func RespondAndLogError(w http.ResponseWriter, r *http.Request, code int, errMsg string, err error) {
	// log original error
	logError(r, errMsg, err)

	if errMsg == "" {
		var ok bool
		if errMsg, ok = codeToGenericMessage[code]; !ok {
			errMsg = defaultErrorMsg
		}
	}
	w.Header().Set(contentTypeHeader, jsonContentType)
	w.WriteHeader(code)
	ReturnJSON(w, r, map[string]interface{}{
		"status":  errStatus,
		"err_msg": errMsg,
	})
}

func logError(r *http.Request, msg string, err error) {
	method := r.Method
	uri := r.RequestURI
	logMessage := fmt.Sprintf("failed [%s] to [%s], msg: %s", method, uri, msg)
	Trail(ERROR, logMessage, err)
}
