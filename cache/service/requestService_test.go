package service

import (
	"fmt"
	"github.com/darkweak/souin/tests"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/darkweak/souin/cache/types"
	"github.com/darkweak/souin/errors"
)

func populateProviderWithFakeData(providers map[string]types.AbstractProviderInterface) {
	for _, provider := range providers {
		provider.Set(tests.DOMAIN+tests.PATH, []byte("testing value is here for "+tests.PATH), tests.GetMatchedURL(tests.DOMAIN+tests.PATH), time.Duration(20)*time.Second)
		for i := 0; i < 25; i++ {
			provider.Set(
				fmt.Sprintf("%s%s/%d", tests.DOMAIN, tests.PATH, i),
				[]byte(fmt.Sprintf("testing value is here for my first init of %s/%d", tests.PATH, i)),
				tests.GetMatchedURL(tests.DOMAIN+tests.PATH),
				time.Duration(20)*time.Second,
			)
		}
	}
}

func mockResponse(path string, method string, body string, code int) *http.Response {
	return &http.Response{
		Status:           "",
		StatusCode:       code,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           make(map[string][]string),
		Body:             ioutil.NopCloser(strings.NewReader(body)),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request: &http.Request{
			Method: method,
			URL: &url.URL{
				Scheme:     "",
				Opaque:     "",
				User:       nil,
				Host:       tests.DOMAIN,
				Path:       path,
				RawPath:    "",
				ForceQuery: false,
				RawQuery:   "",
				Fragment:   "",
			},
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           nil,
			Body:             nil,
			GetBody:          nil,
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            false,
			Host:             tests.DOMAIN,
			Form:             nil,
			PostForm:         nil,
			MultipartForm:    nil,
			Trailer:          nil,
			RemoteAddr:       "",
			RequestURI:       "",
			TLS:              nil,
			Response:         nil,
		},
		TLS: nil,
	}
}

func TestGetKeyFromResponse(t *testing.T) {
	resp := getKeyFromResponse(mockResponse(tests.PATH, http.MethodGet, "", 200), tests.GetMatchedURL(tests.PATH))
	urlCollapsed := tests.DOMAIN + tests.PATH
	if urlCollapsed != resp {
		errors.GenerateError(t, fmt.Sprintf("Key doesn't return %s but %s", urlCollapsed, resp))
	}
}

func shouldNotHaveKey(pathname string, prs map[string]types.AbstractProviderInterface) bool {
	hasKey := true

	for _, pr := range prs {
		r := pr.Get(pathname)
		if 0 < len(r) {
			hasKey = false
		}
	}

	return hasKey
}

func mockRewriteResponse(method string, body string, path string, code int) []byte {
	return RewriteResponse(mockResponse(tests.PATH+path, method, body, code))
}

//func TestKeyShouldBeDeletedOnPost(t *testing.T) {
//	c := tests.MockConfiguration()
//	prs := providers.InitializeProvider(c)
//	populateProviderWithFakeData(prs)
//	res := mockRewriteResponse(http.MethodPost, "My second response", "/1", 201)
//	if len(res) <= 0 {
//		errors.GenerateError(t, "The response should be valid and filled")
//	}
//	time.Sleep(10 * time.Second)
//	if !shouldNotHaveKey(tests.PATH, prs) {
//		errors.GenerateError(t, "The key "+tests.DOMAIN+tests.PATH+" shouldn't exist.")
//	}
//	prs["olric"].Reset()
//}

//func TestRewriteBody(t *testing.T) {
//	response := mockRewriteResponse(http.MethodPost, "My second response", "", 201)
//	if response == nil || len(response) <= 0 {
//		errors.GenerateError(t, "Rewrite body should return an empty response")
//	}
//}
//
//func verifyKeysExists(t *testing.T, path string, keys []string, isKeyDeleted bool, prs map[string]types.AbstractProviderInterface) {
//	time.Sleep(10 * time.Second)
//
//	for _, i := range keys {
//		if !shouldNotHaveKey(tests.PATH+i, prs) == isKeyDeleted {
//			errors.GenerateError(t, "The key "+tests.DOMAIN+path+i+" shouldn't exist.")
//		}
//	}
//}
//
//func TestKeyShouldBeDeletedOnPut(t *testing.T) {
//	c := tests.MockConfiguration()
//	prs := providers.InitializeProvider(c)
//	populateProviderWithFakeData(prs)
//	mockResponse("/1", http.MethodPut, "My second response", 200)
//
//	verifyKeysExists(t, tests.PATH, []string{"", "/1"}, true, prs)
//	prs["olric"].Reset()
//}
//
//func TestKeyShouldBeDeletedOnDelete(t *testing.T) {
//	c := tests.MockConfiguration()
//	prs := providers.InitializeProvider(c)
//	populateProviderWithFakeData(prs)
//	mockResponse("/1", http.MethodDelete, "", 200)
//	verifyKeysExists(t, tests.PATH, []string{"", "/1"}, true, prs)
//	prs["olric"].Reset()
//}
//
//func TestRequestReverseProxy(t *testing.T) {
//	request := httptest.NewRequest("GET", "http://"+tests.DOMAIN+tests.PATH, nil)
//	conf := tests.MockConfiguration()
//	u, _ := url.Parse(conf.GetReverseProxyURL())
//	prs := providers.InitializeProvider(conf)
//	response := RequestReverseProxy(
//		request,
//		&types.RetrieverResponseProperties{
//			Providers:       prs,
//			Configuration:   conf,
//			MatchedURL:      tests.GetMatchedURL(tests.PATH),
//			ReverseProxyURL: u,
//		},
//	)
//	prs["olric"].Reset()
//
//	if response.Proxy == nil || response.Request == nil {
//		errors.GenerateError(t, "Response proxy and request shouldn't be empty")
//	}
//}

func TestCommonLoadingRequest(t *testing.T) {
	body := "My testable response"
	lenBody := len([]byte(body))
	response := responseBodyExtractor(mockResponse(tests.PATH, http.MethodGet, body, 200))

	if "" == string(response) {
		errors.GenerateError(t, "Body shouldn't be empty")
	}
	if body != string(response) || lenBody != len(response) {
		errors.GenerateError(t, fmt.Sprintf("Body %s doesn't match attempted %s", string(response), body))
	}

	body = "Another body with <h1>HTML</h1>"
	lenBody = len([]byte(body))
	response = responseBodyExtractor(mockResponse(tests.PATH, http.MethodGet, body, 200))

	if "" == string(response) {
		errors.GenerateError(t, "Body shouldn't be empty")
	}
	if body != string(response) || lenBody != len(response) {
		errors.GenerateError(t, fmt.Sprintf("Body %s doesn't match attempted %s", string(response), body))
	}

	response = responseBodyExtractor(mockResponse(tests.PATH+"/another", http.MethodGet, body, 200))

	if "" == string(response) {
		errors.GenerateError(t, "Body shouldn't be empty")
	}
	if body != string(response) || lenBody != len(response) {
		errors.GenerateError(t, fmt.Sprintf("Body %s doesn't match attempted %s", string(response), body))
	}
}
