package yaml

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// MergeFiles merges two YAML files, with the second file (overlayPath) taking precedence.
// If overlayPath doesn't exist, only the base file content is returned.
// Returns the merged result as a map[string]interface{}.
func MergeFiles(basePath, overlayPath string) (map[string]interface{}, error) {
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base file: %w", err)
	}

	var base map[string]interface{}
	if err := yaml.Unmarshal(baseData, &base); err != nil {
		return nil, fmt.Errorf("failed to parse base YAML: %w", err)
	}

	// Try to read overlay file, but don't fail if it doesn't exist
	overlayData, err := os.ReadFile(overlayPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read overlay file: %w", err)
		}
		// If overlay file doesn't exist, normalize and return just the base
		return NormalizeMap(base), nil
	}

	var overlay map[string]interface{}
	if err := yaml.Unmarshal(overlayData, &overlay); err != nil {
		return nil, fmt.Errorf("failed to parse overlay YAML: %w", err)
	}

	return MergeMaps(base, overlay), nil
}

// MergeMaps merges two maps, with the src map taking precedence over dst.
// Maps are merged recursively, arrays are merged by name field if present.
func MergeMaps(dst, src map[string]interface{}) map[string]interface{} {
	// Helper function to convert map[interface{}]interface{} to map[string]interface{}
	convertMap := func(m interface{}) (map[string]interface{}, bool) {
		switch mapVal := m.(type) {
		case map[string]interface{}:
			return mapVal, true
		case map[interface{}]interface{}:
			result := make(map[string]interface{})
			for k, v := range mapVal {
				if keyStr, ok := k.(string); ok {
					result[keyStr] = v
				}
			}
			return result, true
		default:
			return nil, false
		}
	}

	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			switch dstValTyped := dstVal.(type) {
			case map[string]interface{}:
				if srcValConverted, canConvert := convertMap(srcVal); canConvert {
					dst[key] = MergeMaps(dstValTyped, srcValConverted)
				} else {
					dst[key] = srcVal
				}
			case []interface{}:
				if srcValTyped, ok := srcVal.([]interface{}); ok {
					dst[key] = MergeListsByName(dstValTyped, srcValTyped)
				} else {
					dst[key] = srcVal
				}
			default:
				// Handle the case where dst value is map[interface{}]interface{}
				if dstValConverted, canConvert := convertMap(dstVal); canConvert {
					if srcValConverted, canConvertSrc := convertMap(srcVal); canConvertSrc {
						dst[key] = MergeMaps(dstValConverted, srcValConverted)
					} else {
						dst[key] = srcVal
					}
				} else {
					dst[key] = srcVal
				}
			}
		} else {
			dst[key] = srcVal
		}
	}
	// Normalize the result to ensure consistent types
	return NormalizeMap(dst)
}

// MergeListsByName merges two arrays by matching items with the same "name" field.
// Items from srcList take precedence over items from dstList with the same name.
// Items without a "name" field or with unique names are appended.
func MergeListsByName(dstList, srcList []interface{}) []interface{} {
	dstMap := make(map[string]map[string]interface{})
	order := []string{}

	// Helper function to convert map[interface{}]interface{} to map[string]interface{}
	convertMap := func(m interface{}) (map[string]interface{}, bool) {
		switch mapVal := m.(type) {
		case map[string]interface{}:
			return mapVal, true
		case map[interface{}]interface{}:
			result := make(map[string]interface{})
			for k, v := range mapVal {
				if keyStr, ok := k.(string); ok {
					result[keyStr] = v
				}
			}
			return result, true
		default:
			return nil, false
		}
	}

	for _, item := range dstList {
		if m, ok := convertMap(item); ok {
			if name, ok := m["name"].(string); ok {
				copy := make(map[string]interface{})
				for k, v := range m {
					copy[k] = v
				}
				dstMap[name] = copy
				order = append(order, name)
			}
		}
	}

	for _, item := range srcList {
		if srcItem, ok := convertMap(item); ok {
			if name, ok := srcItem["name"].(string); ok {
				if dstItem, found := dstMap[name]; found {
					dstMap[name] = MergeMaps(dstItem, srcItem)
				} else {
					copy := make(map[string]interface{})
					for k, v := range srcItem {
						copy[k] = v
					}
					dstMap[name] = copy
					order = append(order, name)
				}
			}
		}
	}

	var result []interface{}
	seen := make(map[string]bool)
	for _, name := range order {
		if !seen[name] {
			result = append(result, dstMap[name])
			seen[name] = true
		}
	}
	return result
}

// NormalizeMap ensures all nested maps are map[string]interface{} instead of map[interface{}]interface{}
func NormalizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = NormalizeValue(v)
	}
	return result
}

// NormalizeValue recursively normalizes values to ensure consistent types
func NormalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			if keyStr, ok := k.(string); ok {
				result[keyStr] = NormalizeValue(v)
			}
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = NormalizeValue(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = NormalizeValue(item)
		}
		return result
	default:
		return v
	}
}
