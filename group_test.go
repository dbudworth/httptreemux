package httptreemux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyGroupAndMapping(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			//everything is good, it paniced
		} else {
			t.Error(`Expected NewGroup("")`)
		}
	}()
	New().GET("", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {})
}
func TestSubGroupSlashMapping(t *testing.T) {
	r := New()
	r.NewGroup("/foo").GET("/", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.WriteHeader(200)
	})

	var req *http.Request
	var recorder *httptest.ResponseRecorder

	req, _ = http.NewRequest("GET", "/foo", nil)
	recorder = httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != 301 { //should get redirected
		t.Error(`/foo on NewGroup("/foo").GET("/") should result in 301 response, got:`, recorder.Code)
	}

	req, _ = http.NewRequest("GET", "/foo/", nil)
	recorder = httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != 200 {
		t.Error(`/foo/ on NewGroup("/foo").GET("/"") should result in 200 response, got:`, recorder.Code)
	}
}

func TestSubGroupEmptyMapping(t *testing.T) {
	r := New()
	r.NewGroup("/foo").GET("", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.WriteHeader(200)
	})
	req, _ := http.NewRequest("GET", "/foo", nil)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != 200 {
		t.Error(`/foo on NewGroup("/foo").GET("") should result in 200 response, got:`, recorder.Code)
	}
}

func TestGroupMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testGroupMethods(t, scenario.RequestCreator)
	}
}

func TestInvalidHandle(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad handle path should have caused a panic")
		}
	}()
	New().NewGroup("/foo").GET("bar", nil)
}

func TestInvalidSubPath(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad sub-path should have caused a panic")
		}
	}()
	New().NewGroup("/foo").NewGroup("bar")
}

func TestInvalidPath(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad path should have caused a panic")
		}
	}()
	New().NewGroup("foo")
}

//Liberally borrowed from router_test
func testGroupMethods(t *testing.T, reqGen RequestCreator) {
	var result string
	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			result = method
		}
	}
	router := New()
	// Testing with a sub-group of a group as that will test everything at once
	g := router.NewGroup("/base").NewGroup("/user")
	g.GET("/:param", makeHandler("GET"))
	g.POST("/:param", makeHandler("POST"))
	g.PATCH("/:param", makeHandler("PATCH"))
	g.PUT("/:param", makeHandler("PUT"))
	g.DELETE("/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		w := httptest.NewRecorder()
		r, _ := reqGen(method, "/base/user/"+method, nil)
		router.ServeHTTP(w, r)
		if expect == "" && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Method %s not expected to match but saw code %d", w.Code)
		}

		if result != expect {
			t.Errorf("Method %s got result %s", method, result)
		}
	}

	testMethod("GET", "GET")
	testMethod("POST", "POST")
	testMethod("PATCH", "PATCH")
	testMethod("PUT", "PUT")
	testMethod("DELETE", "DELETE")
	t.Log("Test HeadCanUseGet = true")
	testMethod("HEAD", "GET")

	router.HeadCanUseGet = false
	t.Log("Test HeadCanUseGet = false")
	testMethod("HEAD", "")

	router.HEAD("/base/user/:param", makeHandler("HEAD"))

	t.Log("Test HeadCanUseGet = false with explicit HEAD handler")
	testMethod("HEAD", "HEAD")
	router.HeadCanUseGet = true
	t.Log("Test HeadCanUseGet = true with explicit HEAD handler")
	testMethod("HEAD", "HEAD")
}
