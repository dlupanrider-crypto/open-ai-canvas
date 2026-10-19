package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunningHubVideoPollRecoversFromTransientGatewayFailure(t *testing.T) {
	allowLoopbackProviderTest(t)
	pollCalls := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/task/openapi/outputs":
			pollCalls++
			w.Header().Set("Content-Type", "application/json")
			if pollCalls == 1 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			_, _ = w.Write([]byte(`{"code":0,"data":[{"fileUrl":"` + server.URL + `/result.mp4"}]}`))
		case "/result.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			_, _ = w.Write([]byte("video"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := providerConfig{BaseURL: server.URL, APIKey: "test-key"}
	ctx := context.Background()
	result, err := (&Service{}).pollRunningHubVideoWorkflowWithPolicy(ctx, config, server.URL, "provider-task-1", fastVideoPollPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result["mode"] != "video" || pollCalls != 2 {
		t.Fatalf("result = %#v, poll calls = %d", result, pollCalls)
	}
}

func TestRunningHubVideoPollHandlesTextStatus(t *testing.T) {
	allowLoopbackProviderTest(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		status := []string{"QUEUED", "RUNNING", "SUCCESS"}[min(calls-1, 2)]
		data := map[string]any{"status": status}
		if status == "SUCCESS" {
			data["results"] = []any{map[string]any{"url": "data:video/mp4;base64,dmlkZW8="}}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
	}))
	defer server.Close()
	result, err := (&Service{}).pollRunningHubVideoWorkflowWithPolicy(context.Background(), providerConfig{APIKey: "test-key"}, server.URL, "task-id", fastVideoPollPolicy())
	if err != nil || result["mode"] != "video" || calls != 3 {
		t.Fatalf("result=%#v calls=%d err=%v", result, calls, err)
	}
}
