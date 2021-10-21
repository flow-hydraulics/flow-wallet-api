package debug

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Service struct {
	RepoUrl   string
	Sha1ver   string
	BuildTime string
}

// Thanks to:
// https://github.com/kjk/go-cookbook/tree/master/embed-build-number

func servePlainText(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s)) // nolint
}

func (d *Service) HandleDebug(w http.ResponseWriter, r *http.Request) {
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
	a = append(a, fmt.Sprintf("ver: %s/commit/%s", d.RepoUrl, d.Sha1ver))
	a = append(a, fmt.Sprintf("built on: %s", d.BuildTime))
	a = append(a, fmt.Sprintf("api version called: %s", v["apiVersion"]))

	s = strings.Join(a, "\n")

	servePlainText(w, s)
}
