package provider

import (
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	colorpb "google.golang.org/genproto/googleapis/type/color"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestColorBlockToProtoUsesFloatChannels(t *testing.T) {
	got, err := colorBlockToProto([]interface{}{
		map[string]interface{}{
			"red":   0.2,
			"green": 0.4,
			"blue":  0.6,
		},
	})
	if err != nil {
		t.Fatalf("colorBlockToProto returned error: %v", err)
	}
	if got == nil {
		t.Fatal("colorBlockToProto returned nil")
	}

	assertFloat32(t, got.Red, 0.2)
	assertFloat32(t, got.Green, 0.4)
	assertFloat32(t, got.Blue, 0.6)
	if got.Alpha == nil {
		t.Fatal("Alpha = nil, want wrapper value")
	}
	assertFloat32(t, got.Alpha.Value, 1)
}

func TestColorBlockToProtoReturnsNilForEmptyBlock(t *testing.T) {
	got, err := colorBlockToProto([]interface{}{})
	if err != nil {
		t.Fatalf("colorBlockToProto returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("colorBlockToProto([]) = %#v, want nil", got)
	}
}

func TestColorBlockToProtoSetsOpaqueAlpha(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"color": colorBlockSchema("The color.", false),
	}, map[string]interface{}{
		"color": []interface{}{
			map[string]interface{}{
				"red":   0.2,
				"green": 0.4,
				"blue":  0.6,
			},
		},
	})

	got, err := colorBlockToProto(resourceData.Get("color").([]interface{}))
	if err != nil {
		t.Fatalf("colorBlockToProto returned error: %v", err)
	}
	if got == nil {
		t.Fatal("colorBlockToProto returned nil")
	}
	if got.Alpha == nil {
		t.Fatal("Alpha = nil, want wrapper value")
	}
	assertFloat32(t, got.Alpha.Value, 1)
}

func TestColorBlockToProtoRejectsOutOfRangeChannels(t *testing.T) {
	for _, tc := range []struct {
		name  string
		block map[string]interface{}
	}{
		{name: "red", block: map[string]interface{}{"red": -0.1, "green": 0.4, "blue": 0.6}},
		{name: "green", block: map[string]interface{}{"red": 0.2, "green": 1.1, "blue": 0.6}},
		{name: "blue", block: map[string]interface{}{"red": 0.2, "green": 0.4, "blue": 1.1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := colorBlockToProto([]interface{}{tc.block}); err == nil {
				t.Fatal("colorBlockToProto returned nil error")
			}
		})
	}
}

func TestProtoColorToBlockReturnsFloatChannels(t *testing.T) {
	got := protoColorToBlock(&colorpb.Color{
		Red:   0.2,
		Green: 0.4,
		Blue:  0.6,
		Alpha: wrapperspb.Float(0.8),
	})
	if len(got) != 1 {
		t.Fatalf("len(protoColorToBlock) = %d, want 1", len(got))
	}
	raw := got[0].(map[string]interface{})
	assertFloat64(t, raw["red"].(float64), 0.2)
	assertFloat64(t, raw["green"].(float64), 0.4)
	assertFloat64(t, raw["blue"].(float64), 0.6)
	if _, ok := raw["alpha"]; ok {
		t.Fatal("protoColorToBlock returned alpha, want RGB-only block")
	}
}

func TestProtoColorToBlockRoundsFloat32Noise(t *testing.T) {
	got := protoColorToBlock(&colorpb.Color{
		Red:   float32(1),
		Green: float32(0.647059),
		Blue:  float32(0),
		Alpha: wrapperspb.Float(1),
	})
	if len(got) != 1 {
		t.Fatalf("len(protoColorToBlock) = %d, want 1", len(got))
	}
	raw := got[0].(map[string]interface{})
	if raw["green"].(float64) != 0.647059 {
		t.Fatalf("green = %.16f, want 0.647059", raw["green"].(float64))
	}
}

func TestProtoColorToBlockReturnsEmptyForNil(t *testing.T) {
	if got := protoColorToBlock(nil); len(got) != 0 {
		t.Fatalf("protoColorToBlock(nil) = %#v, want empty list", got)
	}
}

func TestColorBlockSchemaUsesBlockModeForConfigurableColor(t *testing.T) {
	configurable := colorBlockSchema("The color.", false)
	if configurable.ConfigMode != schema.SchemaConfigModeBlock {
		t.Fatalf("configurable ConfigMode = %v, want SchemaConfigModeBlock", configurable.ConfigMode)
	}

	computed := colorBlockSchema("The color.", true)
	if computed.ConfigMode != schema.SchemaConfigModeAuto {
		t.Fatalf("computed ConfigMode = %v, want SchemaConfigModeAuto", computed.ConfigMode)
	}
}

func assertFloat32(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 0.00001 {
		t.Fatalf("got %f, want %f", got, want)
	}
}

func assertFloat64(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.00001 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
