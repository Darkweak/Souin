package service

import (
	"github.com/darkweak/souin/cache/types"
	"net/http"
	"strings"
)

// ServeResponse serve the response
func ServeResponse(
	res http.ResponseWriter,
	req *http.Request,
	retriever types.RetrieverResponsePropertiesInterface,
	callback func(rw http.ResponseWriter, rq *http.Request, r types.RetrieverResponsePropertiesInterface),
) {
	path := req.Host + req.URL.Path
	regexpURL := retriever.GetRegexpUrls().FindString(path)
	if "" != regexpURL {
		url := retriever.GetConfiguration().GetUrls()[regexpURL]
		retriever.SetMatchedURL(url)
	}
	headers := ""
	if retriever.GetMatchedURL().Headers != nil && len(retriever.GetMatchedURL().Headers) > 0 {
		for _, h := range retriever.GetMatchedURL().Headers {
			headers += strings.ReplaceAll(req.Header.Get(h), " ", "")
		}
	}

	callback(res, req, retriever)
}
