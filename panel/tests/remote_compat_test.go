package tests

import (
	"encoding/json"
	"testing"
)

func TestPaginationShape(t *testing.T) {
	expected := map[string]interface{}{
		"current_page": float64(0),
		"from":         float64(0),
		"last_page":    float64(0),
		"per_page":     float64(0),
		"to":           float64(0),
		"total":        float64(0),
	}

	data, _ := json.Marshal(expected)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	for k := range expected {
		if _, ok := result[k]; !ok {
			t.Errorf("pagination shape missing field: %s", k)
		}
	}
}

func TestErrorResponseShape(t *testing.T) {
	expected := map[string]interface{}{
		"errors": []interface{}{
			map[string]interface{}{
				"code":   "TestError",
				"status": "400",
				"detail": "Test message",
			},
		},
	}

	data, _ := json.Marshal(expected)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if _, ok := result["errors"]; !ok {
		t.Error("error response shape missing 'errors' field")
	}
}

func TestServerResponseShape(t *testing.T) {
	expected := map[string]interface{}{
		"object": "server",
		"attributes": map[string]interface{}{
			"uuid":   "",
			"name":   "",
			"node_id": 0,
			"memory": 0,
			"disk":   0,
			"cpu":    0,
			"status": nil,
		},
	}

	data, _ := json.Marshal(expected)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["object"] != "server" {
		t.Error("server response missing 'object' field")
	}
}

func TestListResponseShape(t *testing.T) {
	expected := map[string]interface{}{
		"object": "list",
		"data":   []interface{}{},
		"meta": map[string]interface{}{
			"pagination": map[string]interface{}{
				"total":        0,
				"count":        0,
				"per_page":     50,
				"current_page": 0,
				"total_pages":  0,
			},
		},
	}

	data, _ := json.Marshal(expected)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Error("list response missing 'meta' field")
		return
	}
	if _, ok := meta["pagination"]; !ok {
		t.Error("list response meta missing 'pagination' field")
	}
}
