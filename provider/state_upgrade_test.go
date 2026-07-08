package provider

import (
	"context"
	"math"
	"testing"
)

func TestUpgradeEnvironmentColorStateConvertsLegacyStringColor(t *testing.T) {
	got, err := upgradeEnvironmentColorState(context.Background(), map[string]interface{}{
		"resource_id": "prod",
		"title":       "Prod",
		"color":       "#ff0000",
	}, nil)
	if err != nil {
		t.Fatalf("upgradeEnvironmentColorState returned error: %v", err)
	}

	assertColorBlock(t, got["color"], 1, 0, 0)
}

func TestUpgradeProjectColorStateConvertsLegacyIssueLabelStringColors(t *testing.T) {
	got, err := upgradeProjectColorState(context.Background(), map[string]interface{}{
		"issue_labels": []interface{}{
			map[string]interface{}{
				"value": "release",
				"color": "#00ff80",
				"group": "type",
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("upgradeProjectColorState returned error: %v", err)
	}

	labels := got["issue_labels"].([]interface{})
	label := labels[0].(map[string]interface{})
	assertColorBlock(t, label["color"], 0, 1, 0.501961)
}

func TestUpgradeSettingColorStateConvertsLegacyEnvironmentStringColors(t *testing.T) {
	got, err := upgradeSettingColorState(context.Background(), map[string]interface{}{
		"name": "settings/ENVIRONMENT",
		"environment_setting": []interface{}{
			map[string]interface{}{
				"environment": []interface{}{
					map[string]interface{}{
						"id":        "prod",
						"title":     "Prod",
						"color":     "336699",
						"protected": true,
					},
				},
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("upgradeSettingColorState returned error: %v", err)
	}

	settings := got["environment_setting"].([]interface{})
	setting := settings[0].(map[string]interface{})
	environments := setting["environment"].([]interface{})
	environment := environments[0].(map[string]interface{})
	assertColorBlock(t, environment["color"], 0.2, 0.4, 0.6)
}

func TestUpgradeSettingColorStatePreservesMaximumRoleExpiration(t *testing.T) {
	got, err := upgradeSettingColorState(context.Background(), map[string]interface{}{
		"name": "settings/WORKSPACE_PROFILE",
		"workspace_profile": []interface{}{
			map[string]interface{}{
				"maximum_role_expiration_in_seconds": 3600,
				"announcement": []interface{}{
					map[string]interface{}{
						"text":  "maintenance",
						"link":  "https://example.com",
						"level": "WARNING",
					},
				},
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("upgradeSettingColorState returned error: %v", err)
	}

	profiles := got["workspace_profile"].([]interface{})
	profile := profiles[0].(map[string]interface{})
	if got := profile["maximum_role_expiration_in_seconds"]; got != 3600 {
		t.Fatalf("maximum_role_expiration_in_seconds = %#v, want 3600", got)
	}
	if _, ok := profile["maximum_request_expiration_in_seconds"]; ok {
		t.Fatal("maximum_request_expiration_in_seconds was added")
	}
	announcement := profile["announcement"].([]interface{})[0].(map[string]interface{})
	if _, ok := announcement["level"]; ok {
		t.Fatal("announcement.level was not removed")
	}
}

func TestUpgradeLegacyColorStringRejectsInvalidColor(t *testing.T) {
	if _, err := upgradeLegacyColorString("not-a-color"); err == nil {
		t.Fatal("upgradeLegacyColorString returned nil error")
	}
}

func assertColorBlock(t *testing.T, raw interface{}, red, green, blue float64) {
	t.Helper()

	blocks, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("color = %#v, want []interface{}", raw)
	}
	if len(blocks) != 1 {
		t.Fatalf("len(color) = %d, want 1", len(blocks))
	}
	block, ok := blocks[0].(map[string]interface{})
	if !ok {
		t.Fatalf("color[0] = %#v, want map[string]interface{}", blocks[0])
	}

	assertClose(t, block["red"].(float64), red)
	assertClose(t, block["green"].(float64), green)
	assertClose(t, block["blue"].(float64), blue)
}

func assertClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
