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

func RespondAndLogError(w http.ResponseWriter, r *http.Request, code int, errMsg string, err error) {
	// log original error
	logError(r, errMsg, err)

	if errMsg == "" {
		if code < 100 || code > 511 {
			code = http.StatusInternalServerError
			errMsg = defaultErrorMsg
		} else {
			errMsg = fmt.Sprintf("%d. %s", code, http.StatusText(code))
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
