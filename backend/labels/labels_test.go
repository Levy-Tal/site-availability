package labels

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeLabels(t *testing.T) {
	t.Run("merge labels with priority order", func(t *testing.T) {
		serverLabels := map[string]string{
			"environment": "production",
			"server":      "serverA",
			"common":      "server-value",
		}

		sourceLabels := map[string]string{
			"instance": "prom1",
			"arc":      "x86",
			"common":   "source-value", // Should override server
		}

		appLabels := map[string]string{
			"cluster": "prod",
			"network": "pvc0213",
			"common":  "app-value", // Should override both server and source
		}

		merged := MergeLabels(serverLabels, sourceLabels, appLabels)

		// Check all labels are present
		assert.Equal(t, "production", merged["environment"]) // From server
		assert.Equal(t, "serverA", merged["server"])         // From server
		assert.Equal(t, "prom1", merged["instance"])         // From source
		assert.Equal(t, "x86", merged["arc"])                // From source
		assert.Equal(t, "prod", merged["cluster"])           // From app
		assert.Equal(t, "pvc0213", merged["network"])        // From app

		// Check priority order: app > source > server
		assert.Equal(t, "app-value", merged["common"]) // App should win

		// Total should be 7 unique keys: environment, server, common, instance, arc, cluster, network
		assert.Len(t, merged, 7)
	})

	t.Run("merge with nil/empty labels", func(t *testing.T) {
		serverLabels := map[string]string{
			"server": "test",
		}

		// Test with nil source and app labels
		merged := MergeLabels(serverLabels, nil, nil)
		assert.Equal(t, "test", merged["server"])
		assert.Len(t, merged, 1)

		// Test with empty maps
		merged = MergeLabels(serverLabels, map[string]string{}, map[string]string{})
		assert.Equal(t, "test", merged["server"])
		assert.Len(t, merged, 1)
	})

	t.Run("merge all empty", func(t *testing.T) {
		merged := MergeLabels(nil, nil, nil)
		assert.Empty(t, merged)
		assert.Len(t, merged, 0)
	})
}

func TestLabelManager(t *testing.T) {
	t.Run("new label manager", func(t *testing.T) {
		lm := NewLabelManager()
		assert.NotNil(t, lm)
		assert.NotNil(t, lm.appsByField)
		assert.Len(t, lm.appsByField, 0)
	})

	t.Run("update app labels and find apps", func(t *testing.T) {
		lm := NewLabelManager()

		// Create test apps with labels
		apps := []AppInfo{
			{
				Name:   "app1",
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
					"cluster":     "east",
				},
			},
			{
				Name:   "app2",
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "staging",
					"cluster":     "west",
				},
			},
			{
				Name:   "app3",
				Source: "prom2",
				Labels: map[string]string{
					"team":        "backend",
					"environment": "production",
					"cluster":     "east",
				},
			},
		}

		lm.UpdateAppLabels(apps)

		// Test finding apps by single label - now uses "labels." prefix
		platformApps := lm.FindAppsByLabel("labels.team", "platform")
		assert.ElementsMatch(t, []string{"prom1:app1", "prom1:app2"}, platformApps)

		prodApps := lm.FindAppsByLabel("labels.environment", "production")
		assert.ElementsMatch(t, []string{"prom1:app1", "prom2:app3"}, prodApps)

		eastApps := lm.FindAppsByLabel("labels.cluster", "east")
		assert.ElementsMatch(t, []string{"prom1:app1", "prom2:app3"}, eastApps)

		// Test finding apps by system fields (no prefix)
		sourceApps := lm.FindAppsByLabel("source", "prom1")
		assert.ElementsMatch(t, []string{"prom1:app1", "prom1:app2"}, sourceApps)

		// Test finding apps by non-existent label
		nonExistent := lm.FindAppsByLabel("nonexistent", "value")
		assert.Empty(t, nonExistent)

		// Test finding apps by non-existent value
		nonExistentValue := lm.FindAppsByLabel("labels.team", "nonexistent")
		assert.Empty(t, nonExistentValue)
	})

	t.Run("find apps by multiple labels", func(t *testing.T) {
		lm := NewLabelManager()

		apps := []AppInfo{
			{
				Name:   "app1",
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
					"cluster":     "east",
				},
			},
			{
				Name:   "app2",
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "staging",
					"cluster":     "east",
				},
			},
			{
				Name:   "app3",
				Source: "prom2",
				Labels: map[string]string{
					"team":        "backend",
					"environment": "production",
					"cluster":     "east",
				},
			},
		}

		lm.UpdateAppLabels(apps)

		// Test AND logic: find apps that are platform AND production
		filters := map[string]string{
			"labels.team":        "platform",
			"labels.environment": "production",
		}
		result := lm.FindAppsByLabels(filters)
		assert.ElementsMatch(t, []string{"prom1:app1"}, result)

		// Test AND logic: find apps in east cluster AND platform team
		filters = map[string]string{
			"labels.team":    "platform",
			"labels.cluster": "east",
		}
		result = lm.FindAppsByLabels(filters)
		assert.ElementsMatch(t, []string{"prom1:app1", "prom1:app2"}, result)

		// Test no matches
		filters = map[string]string{
			"labels.team":        "platform",
			"labels.environment": "production",
			"labels.cluster":     "west", // No apps match this combination
		}
		result = lm.FindAppsByLabels(filters)
		assert.Empty(t, result)

		// Test empty filters
		result = lm.FindAppsByLabels(map[string]string{})
		assert.Empty(t, result)
	})

	t.Run("get label keys and values", func(t *testing.T) {
		lm := NewLabelManager()

		apps := []AppInfo{
			{
				Name:   "app1",
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
				},
			},
			{
				Name:   "app2",
				Source: "prom2",
				Labels: map[string]string{
					"team":        "backend",
					"environment": "staging",
					"cluster":     "east",
				},
			},
		}

		lm.UpdateAppLabels(apps)

		// Test getting all field keys (now includes both system fields and labels with prefix)
		keys := lm.GetLabelKeys()
		expectedKeys := []string{"name", "source", "labels.team", "labels.environment", "labels.cluster"}
		assert.ElementsMatch(t, expectedKeys, keys)

		// Test getting values for specific label keys (with prefix)
		teamValues := lm.GetLabelValues("labels.team")
		assert.ElementsMatch(t, []string{"platform", "backend"}, teamValues)

		envValues := lm.GetLabelValues("labels.environment")
		assert.ElementsMatch(t, []string{"production", "staging"}, envValues)

		clusterValues := lm.GetLabelValues("labels.cluster")
		assert.ElementsMatch(t, []string{"east"}, clusterValues)

		// Test getting values for system fields (without prefix)
		sourceValues := lm.GetLabelValues("source")
		assert.ElementsMatch(t, []string{"prom1", "prom2"}, sourceValues)

		nameValues := lm.GetLabelValues("name")
		assert.ElementsMatch(t, []string{"app1", "app2"}, nameValues)

		// Test non-existent key
		nonExistent := lm.GetLabelValues("nonexistent")
		assert.Empty(t, nonExistent)
	})

	t.Run("get stats", func(t *testing.T) {
		lm := NewLabelManager()

		apps := []AppInfo{
			{
				Name:   "app1",
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
				},
			},
			{
				Name:   "app2",
				Source: "prom2",
				Labels: map[string]string{
					"team":        "platform", // Same value as app1
					"environment": "staging",
				},
			},
		}

		lm.UpdateAppLabels(apps)

	})

	t.Run("update with empty apps", func(t *testing.T) {
		lm := NewLabelManager()

		// First add some apps
		apps := []AppInfo{
			{
				Name:   "app1",
				Source: "prom1",
				Labels: map[string]string{
					"team": "platform",
				},
			},
		}
		lm.UpdateAppLabels(apps)

		// Verify it was added (name, source + labels.team = 3 fields)
		assert.Len(t, lm.GetLabelKeys(), 3)

		// Now update with empty list (should clear everything)
		lm.UpdateAppLabels([]AppInfo{})

		// Should be empty now
		assert.Empty(t, lm.GetLabelKeys())
		assert.Empty(t, lm.FindAppsByLabel("labels.team", "platform"))
	})

	t.Run("apps without labels", func(t *testing.T) {
		lm := NewLabelManager()

		apps := []AppInfo{
			{
				Name:   "app1",
				Source: "prom1",
				Labels: nil, // Explicitly nil
			},
			{
				Name:   "app2",
				Source: "prom2",
				Labels: map[string]string{}, // Empty map
			},
		}

		lm.UpdateAppLabels(apps)

		// Should handle gracefully and still index system fields
		keys := lm.GetLabelKeys()
		// Even without user labels, we should have system fields: name, source
		expectedKeys := []string{"name", "source"}
		assert.ElementsMatch(t, expectedKeys, keys)
	})

	t.Run("apps with same name from different sources", func(t *testing.T) {
		lm := NewLabelManager()

		// This is the critical test case: apps with same name but different sources
		apps := []AppInfo{
			{
				Name:   "app1", // Same name
				Source: "prom1",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "production",
				},
			},
			{
				Name:   "app1", // Same name, different source
				Source: "prom2",
				Labels: map[string]string{
					"team":        "backend",
					"environment": "staging",
				},
			},
			{
				Name:   "app1", // Same name, different source and labels
				Source: "site",
				Labels: map[string]string{
					"team":        "platform",
					"environment": "development",
				},
			},
		}

		lm.UpdateAppLabels(apps)

		// Test that all three apps are properly indexed despite having the same name
		platformApps := lm.FindAppsByLabel("labels.team", "platform")
		assert.ElementsMatch(t, []string{"prom1:app1", "site:app1"}, platformApps)

		backendApps := lm.FindAppsByLabel("labels.team", "backend")
		assert.ElementsMatch(t, []string{"prom2:app1"}, backendApps)

		prodApps := lm.FindAppsByLabel("labels.environment", "production")
		assert.ElementsMatch(t, []string{"prom1:app1"}, prodApps)

		stagingApps := lm.FindAppsByLabel("labels.environment", "staging")
		assert.ElementsMatch(t, []string{"prom2:app1"}, stagingApps)

		devApps := lm.FindAppsByLabel("labels.environment", "development")
		assert.ElementsMatch(t, []string{"site:app1"}, devApps)

		// Test multi-label filtering
		platformProdApps := lm.FindAppsByLabels(map[string]string{
			"labels.team":        "platform",
			"labels.environment": "production",
		})
		assert.ElementsMatch(t, []string{"prom1:app1"}, platformProdApps)

		// Verify stats are correct
	})
}

func TestIntersectSlices(t *testing.T) {
	t.Run("normal intersection", func(t *testing.T) {
		slice1 := []string{"app1", "app2", "app3"}
		slice2 := []string{"app2", "app3", "app4"}

		result := intersectSlices(slice1, slice2)
		assert.ElementsMatch(t, []string{"app2", "app3"}, result)
	})

	t.Run("no intersection", func(t *testing.T) {
		slice1 := []string{"app1", "app2"}
		slice2 := []string{"app3", "app4"}

		result := intersectSlices(slice1, slice2)
		assert.Empty(t, result)
	})

	t.Run("empty slices", func(t *testing.T) {
		result := intersectSlices([]string{}, []string{"app1"})
		assert.Empty(t, result)

		result = intersectSlices([]string{"app1"}, []string{})
		assert.Empty(t, result)

		result = intersectSlices([]string{}, []string{})
		assert.Empty(t, result)
	})

	t.Run("identical slices", func(t *testing.T) {
		slice := []string{"app1", "app2"}
		result := intersectSlices(slice, slice)
		assert.ElementsMatch(t, []string{"app1", "app2"}, result)
	})

	t.Run("duplicate values handled correctly", func(t *testing.T) {
		slice1 := []string{"app1", "app1", "app2"}
		slice2 := []string{"app1", "app2", "app2"}

		result := intersectSlices(slice1, slice2)
		// Should not have duplicates in result
		assert.ElementsMatch(t, []string{"app1", "app2"}, result)
	})
}

// Benchmark tests for performance validation
func BenchmarkMergeLabels(b *testing.B) {
	serverLabels := map[string]string{
		"server": "serverA",
		"env":    "prod",
	}
	sourceLabels := map[string]string{
		"instance": "prom1",
		"arc":      "x86",
	}
	appLabels := map[string]string{
		"cluster": "east",
		"network": "vpc123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeLabels(serverLabels, sourceLabels, appLabels)
	}
}

func BenchmarkLabelManagerLookup(b *testing.B) {
	lm := NewLabelManager()

	// Create 1000 apps with various labels
	apps := make([]AppInfo, 1000)
	for i := 0; i < 1000; i++ {
		apps[i] = AppInfo{
			Name:   fmt.Sprintf("app%d", i),
			Source: fmt.Sprintf("prom%d", i%5), // 5 different sources
			Labels: map[string]string{
				"team":        fmt.Sprintf("team%d", i%10),   // 10 different teams
				"environment": fmt.Sprintf("env%d", i%3),     // 3 different environments
				"cluster":     fmt.Sprintf("cluster%d", i%5), // 5 different clusters
			},
		}
	}

	lm.UpdateAppLabels(apps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.FindAppsByLabel("team", "team5")
	}
}
