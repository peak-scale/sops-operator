/*
Copyright 2024 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package decryptor

import (
	"encoding/json"
	"testing"
)

func TestConvertInterfaceMapToStringMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map returns empty map",
			input:    map[string]interface{}{},
			expected: map[string]string{},
		},
		{
			name: "string values preserved",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "integer values converted to string (type:int fix)",
			input: map[string]interface{}{
				"port":     8080,
				"replicas": 3,
			},
			expected: map[string]string{
				"port":     "8080",
				"replicas": "3",
			},
		},
		{
			name: "float values converted to string (type:float fix)",
			input: map[string]interface{}{
				"ratio":   0.5,
				"percent": 99.9,
			},
			expected: map[string]string{
				"ratio":   "0.5",
				"percent": "99.9",
			},
		},
		{
			name: "boolean values converted to string (type:bool fix)",
			input: map[string]interface{}{
				"enabled":  true,
				"disabled": false,
			},
			expected: map[string]string{
				"enabled":  "true",
				"disabled": "false",
			},
		},
		{
			name: "mixed types all converted correctly",
			input: map[string]interface{}{
				"name":     "my-service",
				"port":     8080,
				"enabled":  true,
				"ratio":    0.75,
				"password": "secret123",
			},
			expected: map[string]string{
				"name":     "my-service",
				"port":     "8080",
				"enabled":  "true",
				"ratio":    "0.75",
				"password": "secret123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertInterfaceMapToStringMap(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("key %q: expected %q, got %q", k, v, result[k])
				}
			}
		})
	}
}

// TestSopsSecretRawUnmarshal tests that the intermediate struct can handle
// non-string values that SOPS outputs when using type annotations.
// This is the root cause of issue #335.
func TestSopsSecretRawUnmarshal(t *testing.T) {
	// This JSON simulates what SOPS outputs when decrypting values with type:int
	// Note: port value is 8080 (unquoted integer), not "8080" (string)
	jsonWithTypeInt := `{
		"spec": {
			"secrets": [{
				"name": "test-secret",
				"stringData": {
					"port": 8080,
					"host": "localhost",
					"enabled": true,
					"ratio": 0.5
				},
				"data": {
					"count": 42
				}
			}]
		}
	}`

	var target sopsSecretRaw
	err := json.Unmarshal([]byte(jsonWithTypeInt), &target)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON with type:int values: %v", err)
	}

	if len(target.Spec.Secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(target.Spec.Secrets))
	}

	secret := target.Spec.Secrets[0]

	// Verify StringData can hold non-string types
	stringData := convertInterfaceMapToStringMap(secret.StringData)
	expectedStringData := map[string]string{
		"port":    "8080",
		"host":    "localhost",
		"enabled": "true",
		"ratio":   "0.5",
	}

	for k, expected := range expectedStringData {
		if stringData[k] != expected {
			t.Errorf("stringData[%q]: expected %q, got %q", k, expected, stringData[k])
		}
	}

	// Verify Data can hold non-string types
	data := convertInterfaceMapToStringMap(secret.Data)
	if data["count"] != "42" {
		t.Errorf("data[count]: expected \"42\", got %q", data["count"])
	}
}

// TestOriginalStructFailsWithTypeInt demonstrates the original bug - trying to
// unmarshal SOPS output with type:int into map[string]string fails.
func TestOriginalStructFailsWithTypeInt(t *testing.T) {
	// This is what the original code tried to do - unmarshal into map[string]string
	type originalStringData struct {
		StringData map[string]string `json:"stringData"`
	}

	jsonWithTypeInt := `{"stringData": {"port": 8080}}`

	var target originalStringData
	err := json.Unmarshal([]byte(jsonWithTypeInt), &target)

	// This SHOULD fail because Go cannot unmarshal a JSON number into a string
	if err == nil {
		t.Error("expected unmarshal to fail when JSON number is assigned to string field, but it succeeded")
	}
}
