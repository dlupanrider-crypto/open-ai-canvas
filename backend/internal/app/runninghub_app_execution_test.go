package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func runningHubAppTestFields(t *testing.T) []WorkflowField {
	t.Helper()
	// 取自 H3 App 的公开字段顺序与空槽约定，不保留示例提示词或素材。
	raw := []any{
		map[string]any{"nodeId": "51", "fieldName": "image", "fieldType": "IMAGE", "fieldValue": "example.png", "description": "image1（图1）", "fieldData": `[["example.png","None"],{"image_upload":true}]`},
		map[string]any{"nodeId": "49", "fieldName": "image", "fieldType": "IMAGE", "fieldValue": "example2.png"},
		map[string]any{"nodeId": "48", "fieldName": "audio", "fieldType": "AUDIO", "fieldValue": "None"},
		map[string]any{"nodeId": "27", "fieldName": "video", "fieldType": "VIDEO", "fieldValue": ""},
		map[string]any{"nodeId": "252", "fieldName": "aspect_ratio", "fieldType": "LIST", "fieldValue": "16:9 (Widescreen)", "fieldData": `["COMBO",{"options":["1:1 (Square)","16:9 (Widescreen)"]}]`},
		map[string]any{"nodeId": "252", "fieldName": "megapixels", "fieldType": "FLOAT", "fieldValue": "0.5", "fieldData": `["FLOAT",{"min":0.1,"max":16,"step":0.1}]`},
		map[string]any{"nodeId": "263", "fieldName": "text", "fieldType": "STRING", "fieldValue": "example prompt"},
		map[string]any{"nodeId": "43", "fieldName": "image", "fieldType": "IMAGE", "fieldValue": "None"},
	}
	encoded, err := json.Marshal(normalizeManagementAppFields(raw, "video"))
	if err != nil {
		t.Fatal(err)
	}
	var fields []WorkflowField
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	return fields
}

func TestRunningHubAppPublicFieldContract(t *testing.T) {
	fields := runningHubAppTestFields(t)
	var ids []string
	for _, field := range fields {
		ids = append(ids, field.NodeID)
		if field.Required {
			t.Fatalf("invented required field: %#v", field)
		}
	}
	if !reflect.DeepEqual(ids, []string{"51", "49", "48", "27", "252", "252", "263", "43"}) {
		t.Fatalf("order = %v", ids)
	}
	if fields[0].ImageOrder != 1 || fields[1].ImageOrder != 2 || fields[7].ImageOrder != 3 {
		t.Fatal("media slots do not follow public order")
	}
	if fields[0].Label != "image1（图1）" || len(fields[0].Options) != 0 {
		t.Fatalf("media schema = %#v", fields[0])
	}
	if !reflect.DeepEqual(fields[4].Options, []any{"1:1 (Square)", "16:9 (Widescreen)"}) {
		t.Fatalf("options = %#v", fields[4].Options)
	}
	if fields[5].Min != 0.1 || fields[5].Max != float64(16) || fields[5].Step != 0.1 {
		t.Fatalf("range = %#v", fields[5])
	}
	if err := validateRunningHubWorkflowFieldValue(fields[5], "0.55"); err == nil {
		t.Fatal("invalid step accepted")
	}
	if err := validateRunningHubWorkflowFieldValue(fields[4], "9:16"); err == nil {
		t.Fatal("invalid enum accepted")
	}
}

func TestRunningHubAppEmptySlotsAndIndependentValues(t *testing.T) {
	fields := runningHubAppTestFields(t)
	input := canvasGenerationInput{Mode: "video", Config: providerConfig{WebappID: "app-id", Size: "9:16"}, ReferenceImages: []providerMedia{{ID: "image-1"}}}
	items, err := runningHubNodeInfo(fields, map[string]string{"image-1": "api/upload.png"}, input)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api/upload.png", "", "None", "", "16:9 (Widescreen)", "0.5", "", "None"}
	if len(items) != len(want) {
		t.Fatalf("missing public slots: %#v", items)
	}
	for i, value := range want {
		if items[i]["fieldValue"] != value {
			t.Fatalf("item %d = %#v, want %q", i, items[i], value)
		}
	}
	fields[1].Required = true
	if _, err := runningHubNodeInfo(fields, map[string]string{"image-1": "api/upload.png"}, input); err == nil {
		t.Fatal("required media accepted without input")
	}
	fields[1].Enabled = ptr(false)
	if items, err := runningHubNodeInfo(fields, map[string]string{"image-1": "api/upload.png"}, input); err != nil || len(items) != len(want)-1 {
		t.Fatalf("disabled field: %#v, %v", items, err)
	}
}

func TestRunningHubWorkflowMissingMediaStillOmitted(t *testing.T) {
	fields := []WorkflowField{{NodeID: "1", FieldName: "image", Source: "referenceImage"}}
	items, err := runningHubNodeInfo(fields, nil, canvasGenerationInput{})
	if err != nil || len(items) != 0 {
		t.Fatalf("workflow changed: %#v, %v", items, err)
	}
}

func TestRunningHubAppExplicitEmptyZeroAndFalse(t *testing.T) {
	fields := []WorkflowField{
		{NodeID: "1", FieldName: "negative_prompt", FieldValue: ""},
		{NodeID: "2", FieldName: "count", FieldValue: 0},
		{NodeID: "3", FieldName: "enabled", FieldValue: false},
	}
	items, err := runningHubNodeInfo(fields, nil, canvasGenerationInput{Config: providerConfig{WebappID: "app-id"}})
	if err != nil || len(items) != 3 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	for i, want := range []string{"", "0", "false"} {
		if items[i]["fieldValue"] != want {
			t.Fatalf("item %d=%#v, want %q", i, items[i], want)
		}
	}
}

func TestRunningHubAppLegacyEnumFieldData(t *testing.T) {
	field := map[string]any{"fieldType": "LIST", "fieldData": `[{"name":"Fast","index":"fast","description":"快速"},{"name":"Quality","index":"quality"},{"default":"fast"}]`}
	applyManagementAppFieldData(field)
	options := field["options"].([]any)
	if len(options) != 2 || workflowOptionString(options[0]) != "fast" || workflowOptionString(options[1]) != "quality" {
		t.Fatalf("options = %#v", options)
	}
}

func TestRunningHubAppSubmissionAcceptsActiveTask(t *testing.T) {
	allowLoopbackProviderTest(t)
	for _, status := range []string{"RUNNING", "QUEUED"} {
		t.Run(status, func(t *testing.T) {
			var server *httptest.Server
			submissions, polls := 0, 0
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/task/openapi/ai-app/run":
					submissions++
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["webappId"] != "app-id" || body["workflowId"] != nil {
						t.Errorf("invalid app body: %#v", body)
					}
					items, _ := body["nodeInfoList"].([]any)
					if len(items) != 8 || items[0].(map[string]any)["fieldValue"] != "None" || items[2].(map[string]any)["fieldValue"] != "None" {
						t.Errorf("invalid empty slots: %#v", items)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{"taskId": "task-id", "taskStatus": status}})
				case "/task/openapi/outputs":
					polls++
					_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": []any{map[string]any{"fileUrl": server.URL + "/result.mp4"}}})
				case "/result.mp4":
					w.Header().Set("Content-Type", "video/mp4")
					_, _ = w.Write([]byte("test video"))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			result, err := (&Service{}).runRunningHubWorkflow(context.Background(), canvasGenerationInput{Mode: "video", Prompt: "test", Config: providerConfig{BaseURL: server.URL, APIKey: "test-key", WebappID: "app-id", WorkflowFields: runningHubAppTestFields(t)}})
			if err != nil {
				t.Fatal(err)
			}
			if result["mode"] != "video" || submissions != 1 || polls != 1 {
				t.Fatalf("result=%#v submissions=%d polls=%d", result, submissions, polls)
			}
		})
	}
}

func TestRunningHubAppSubmissionRejectsFailure(t *testing.T) {
	allowLoopbackProviderTest(t)
	for _, response := range []string{
		`{"code":401,"msg":"unauthorized","data":{"taskId":"task-id","taskStatus":"RUNNING"}}`,
		`{"code":0,"data":{"taskId":"task-id","taskStatus":"FAILED"}}`,
		`{"code":0,"data":{"taskStatus":"QUEUED"}}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/task/openapi/ai-app/run" {
				t.Errorf("failure must not poll: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(response))
		}))
		_, err := (&Service{}).runRunningHubWorkflow(context.Background(), canvasGenerationInput{Mode: "video", Config: providerConfig{BaseURL: server.URL, APIKey: "test-key", WebappID: "app-id"}})
		server.Close()
		if err == nil {
			t.Fatalf("accepted failure: %s", response)
		}
	}
}

func TestRunningHubAppStatusParsing(t *testing.T) {
	for _, tc := range []struct {
		status string
		code   int
		valid  bool
	}{
		{"RUNNING", 804, true}, {"QUEUED", 813, true}, {"FAILED", 805, true},
		{"SUCCESS", 0, true}, {"0", 0, true}, {"804", 804, true}, {"unknown", 0, false},
	} {
		code, valid := runningHubStatusCode(tc.status)
		if code != tc.code || valid != tc.valid {
			t.Errorf("status %s = %d,%v; want %d,%v", tc.status, code, valid, tc.code, tc.valid)
		}
	}
	for _, value := range []any{"", "invalid", "0.5", "-1", nil} {
		if _, valid := runningHubCode(value); valid {
			t.Errorf("accepted non-code %#v", value)
		}
	}
}
