package yaml

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		dst      map[string]interface{}
		src      map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "merge simple maps",
			dst: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			src: map[string]interface{}{
				"b": 3,
				"c": 4,
			},
			expected: map[string]interface{}{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
		{
			name: "merge nested maps",
			dst: map[string]interface{}{
				"server": map[string]interface{}{
					"port": 8080,
					"host": "localhost",
				},
			},
			src: map[string]interface{}{
				"server": map[string]interface{}{
					"port": 9090,
					"ssl":  true,
				},
			},
			expected: map[string]interface{}{
				"server": map[string]interface{}{
					"port": 9090,
					"host": "localhost",
					"ssl":  true,
				},
			},
		},
		{
			name: "merge arrays with name field",
			dst: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name":  "item1",
						"value": 1,
					},
					map[string]interface{}{
						"name":  "item2",
						"value": 2,
					},
				},
			},
			src: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name":  "item1",
						"value": 10,
					},
					map[string]interface{}{
						"name":  "item3",
						"value": 3,
					},
				},
			},
			expected: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name":  "item1",
						"value": 10,
					},
					map[string]interface{}{
						"name":  "item2",
						"value": 2,
					},
					map[string]interface{}{
						"name":  "item3",
						"value": 3,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeMaps(tt.dst, tt.src)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MergeMaps() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMergeListsByName(t *testing.T) {
	tests := []struct {
		name     string
		dstList  []interface{}
		srcList  []interface{}
		expected []interface{}
	}{
		{
			name: "merge lists with same names",
			dstList: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 1,
				},
			},
			srcList: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 10,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 10,
				},
			},
		},
		{
			name: "merge lists with different names",
			dstList: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 1,
				},
			},
			srcList: []interface{}{
				map[string]interface{}{
					"name":  "item2",
					"value": 2,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 1,
				},
				map[string]interface{}{
					"name":  "item2",
					"value": 2,
				},
			},
		},
		{
			name: "handle items without name field",
			dstList: []interface{}{
				map[string]interface{}{
					"value": 1,
				},
			},
			srcList: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 2,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":  "item1",
					"value": 2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeListsByName(tt.dstList, tt.srcList)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MergeListsByName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMergeFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create test config file
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `server_settings:
  port: "8080"
  host: "localhost"
locations:
  - name: "loc1"
    latitude: 40.7128
    longitude: -74.0060
sources:
  - name: "src1"
    type: "http"
    config:
      url: "https://example.com"`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create test overlay file
	overlayFile := filepath.Join(tempDir, "overlay.yaml")
	overlayContent := `server_settings:
  port: "9090"
  token: "secret-token"
sources:
  - name: "src1"
    config:
      timeout: 30
  - name: "src2"
    type: "tcp"
    config:
      host: "example.com"`
	err = os.WriteFile(overlayFile, []byte(overlayContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test successful merge
	t.Run("successful merge", func(t *testing.T) {
		result, err := MergeFiles(configFile, overlayFile)
		if err != nil {
			t.Errorf("MergeFiles() error = %v", err)
			return
		}

		// Check that server settings were merged
		serverSettings, ok := result["server_settings"].(map[string]interface{})
		if !ok {
			t.Error("server_settings not found or wrong type")
			return
		}

		if serverSettings["port"] != "9090" {
			t.Errorf("Expected port to be overridden to 9090, got %v", serverSettings["port"])
		}

		if serverSettings["host"] != "localhost" {
			t.Errorf("Expected host to be preserved as localhost, got %v", serverSettings["host"])
		}

		if serverSettings["token"] != "secret-token" {
			t.Errorf("Expected token to be added, got %v", serverSettings["token"])
		}

		// Check that sources were merged by name
		sources, ok := result["sources"].([]interface{})
		if !ok {
			t.Error("sources not found or wrong type")
			return
		}

		if len(sources) != 2 {
			t.Errorf("Expected 2 sources, got %d", len(sources))
		}

		// Verify source1 was properly merged
		found := false
		for _, source := range sources {
			sourceMap, ok := source.(map[string]interface{})
			if !ok {
				t.Error("source not a map")
				continue
			}
			if sourceMap["name"] == "src1" {
				found = true
				config, ok := sourceMap["config"].(map[string]interface{})
				if !ok {
					t.Error("source config not a map")
					continue
				}
				if config["url"] != "https://example.com" {
					t.Errorf("Expected url https://example.com, got %v", config["url"])
				}
				if config["timeout"] != 30 {
					t.Errorf("Expected timeout 30, got %v", config["timeout"])
				}
			}
		}
		if !found {
			t.Error("src1 source not found after merge")
		}
	})

	// Test with non-existent base file
	t.Run("non-existent base file", func(t *testing.T) {
		_, err := MergeFiles("non-existent.yaml", overlayFile)
		if err == nil {
			t.Error("Expected error for non-existent base file")
		}
	})

	// Test with non-existent overlay file
	t.Run("non-existent overlay file", func(t *testing.T) {
		result, err := MergeFiles(configFile, "non-existent.yaml")
		if err != nil {
			t.Errorf("Expected success when overlay file doesn't exist, got error: %v", err)
			return
		}

		// Should return just the base file contents
		serverSettings, ok := result["server_settings"].(map[string]interface{})
		if !ok {
			t.Error("server_settings not found or wrong type")
			return
		}

		if serverSettings["port"] != "8080" {
			t.Errorf("Expected port to be 8080, got %v", serverSettings["port"])
		}

		if serverSettings["host"] != "localhost" {
			t.Errorf("Expected host to be localhost, got %v", serverSettings["host"])
		}

		// Token should not be present since overlay file doesn't exist
		if serverSettings["token"] != nil {
			t.Errorf("Expected no token, got %v", serverSettings["token"])
		}
	})

	// Test with invalid YAML in base file
	t.Run("invalid base YAML", func(t *testing.T) {
		invalidBaseFile := filepath.Join(tempDir, "invalid-base.yaml")
		err := os.WriteFile(invalidBaseFile, []byte("invalid: yaml: :"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		_, err = MergeFiles(invalidBaseFile, overlayFile)
		if err == nil {
			t.Error("Expected error for invalid base YAML")
		}
	})

	// Test with invalid YAML in overlay file
	t.Run("invalid overlay YAML", func(t *testing.T) {
		invalidOverlayFile := filepath.Join(tempDir, "invalid-overlay.yaml")
		err := os.WriteFile(invalidOverlayFile, []byte("invalid: yaml: :"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		_, err = MergeFiles(configFile, invalidOverlayFile)
		if err == nil {
			t.Error("Expected error for invalid overlay YAML")
		}
	})
}

func TestNormalizeMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "already normalized map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"nested": "value",
				},
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"nested": "value",
				},
			},
		},
		{
			name: "normalize nested interface{} map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": map[interface{}]interface{}{
					"nested": "value",
				},
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"nested": "value",
				},
			},
		},
		{
			name: "normalize array with maps",
			input: map[string]interface{}{
				"items": []interface{}{
					map[interface{}]interface{}{
						"name": "item1",
					},
					map[string]interface{}{
						"name": "item2",
					},
				},
			},
			expected: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name": "item1",
					},
					map[string]interface{}{
						"name": "item2",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeMap(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NormalizeMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "int value",
			input:    42,
			expected: 42,
		},
		{
			name: "map[interface{}]interface{}",
			input: map[interface{}]interface{}{
				"key": "value",
			},
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "map[string]interface{}",
			input: map[string]interface{}{
				"key": "value",
			},
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "array with mixed types",
			input: []interface{}{
				"string",
				42,
				map[interface{}]interface{}{
					"key": "value",
				},
			},
			expected: []interface{}{
				"string",
				42,
				map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeValue(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NormalizeValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}
