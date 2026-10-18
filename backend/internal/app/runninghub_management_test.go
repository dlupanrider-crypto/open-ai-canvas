package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchRunningHubAppInfoGETProtocol(t *testing.T) {
	allowLoopbackProviderTest(t)
	for _, code := range []int{0, 401} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != http.MethodGet || r.URL.Path != "/api/webapp/apiCallDemo" {
					t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
				}
				if r.URL.Query().Get("apiKey") != "test+key&value" || r.URL.Query().Get("webappId") != "app-id" {
					t.Error("query parameters do not match")
				}
				if r.Header.Get("Authorization") != "Bearer test+key&value" {
					t.Error("missing bearer authorization")
				}
				body, _ := io.ReadAll(r.Body)
				if len(body) != 0 {
					t.Error("GET request must not have a body")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{"code":%d,"msg":"test","data":{"nodeInfoList":[{"nodeId":"39","fieldName":"image","fieldType":"IMAGE","fieldValue":"input.png"}]}}`, code)
			}))
			defer server.Close()
			result, err := (&Service{}).FetchRunningHubAppInfo(context.Background(), RunningHubWorkflowFetchRequest{
				BaseURL: server.URL + "/openapi/v2", APIKey: "test+key&value", WebappID: " app-id ", Capability: "image",
			})
			if !called {
				t.Fatal("upstream was not called")
			}
			if code != 0 {
				if err == nil {
					t.Fatal("upstream business failure must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			fields, ok := result["fields"].([]map[string]any)
			if !ok || len(fields) != 1 || fields[0]["id"] != "39::image" || result["webappId"] != "app-id" {
				t.Fatalf("unexpected normalized result: %#v", result)
			}
		})
	}
}
