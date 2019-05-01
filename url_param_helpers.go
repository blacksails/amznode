package amznode

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi"
)

const validNameRegexpStr = `^[a-zA-Z\d-_]+$`

var validNameRegexp = regexp.MustCompile(validNameRegexpStr)
var errInvalidName = fmt.Errorf("name must match the regex /%s/", validNameRegexpStr)
var errInvalidID = errors.New("ids must be greater than or equal 0")

func urlOrQueryParam(r *http.Request, paramName string) string {
	paramStr := chi.URLParam(r, paramName)
	if paramStr == "" {
		return r.URL.Query().Get(paramName)
	}
	return paramStr
}

func urlParamID(r *http.Request, paramName string) (int, error) {
	idStr := urlOrQueryParam(r, paramName)
	if idStr == "" {
		return 0, nil
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, err
	}
	if !validID(id) {
		return id, errInvalidID
	}
	return id, err
}

func urlParamName(r *http.Request, paramName string) (string, error) {
	name := urlOrQueryParam(r, paramName)
	if !validName(name) {
		return name, errInvalidName
	}
	return name, nil
}

func validName(name string) bool {
	return validNameRegexp.MatchString(name)
}

func validID(id int) bool {
	return id >= 0
}
