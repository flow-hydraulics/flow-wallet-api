package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Thanks to:
// https://github.com/kjk/go-cookbook/tree/master/embed-build-number

func Debug(repoURL, sha1ver, buildtime string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		s := fmt.Sprintf("url: %s %s", r.Method, r.RequestURI)
		a := []string{s}

		a = append(a, "Headers:")
		for k, v := range r.Header {
			if len(v) == 0 {
				a = append(a, k)
			} else if len(v) == 1 {
				s = fmt.Sprintf("  %s: %v", k, v[0])
				a = append(a, s)
			} else {
				a = append(a, "  "+k+":")
				for _, v2 := range v {
					a = append(a, "    "+v2)
				}
			}
		}

		a = append(a, "")
		a = append(a, fmt.Sprintf("ver: %s/commit/%s", repoURL, sha1ver))
		a = append(a, fmt.Sprintf("built on: %s", buildtime))
		a = append(a, fmt.Sprintf("api version called: %s", v["apiVersion"]))

		s = strings.Join(a, "\n")

		servePlainText(rw, s)
	})
}
