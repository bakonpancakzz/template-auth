package tools

import "net/http"

type MiddlewareFunc func(w http.ResponseWriter, r *http.Request) bool

func Chain(h http.HandlerFunc, mw ...MiddlewareFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// global middleware injection :)
		if !UseServer(w, r) {
			return
		}
		if !UseCORS(w, r) {
			return
		}

		for i := 0; i < len(mw); i++ {
			if !mw[i](w, r) {
				return
			}
		}
		h(w, r)
	}
}

type MethodHandler map[string]http.HandlerFunc

func (mh MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := mh[r.Method]; ok {
		handler(w, r)
	} else {
		SendClientError(w, r, ERROR_GENERIC_METHOD_NOT_ALLOWED)
	}
}
