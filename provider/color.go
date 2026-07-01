package provider

import (
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
	colorpb "google.golang.org/genproto/googleapis/type/color"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func colorBlockToProto(rawList []interface{}) (*colorpb.Color, error) {
	if len(rawList) == 0 {
		return nil, nil
	}
	raw, ok := rawList[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid color block")
	}

	red, err := colorChannel(raw, "red")
	if err != nil {
		return nil, err
	}
	green, err := colorChannel(raw, "green")
	if err != nil {
		return nil, err
	}
	blue, err := colorChannel(raw, "blue")
	if err != nil {
		return nil, err
	}

	color := &colorpb.Color{
		Red:   float32(red),
		Green: float32(green),
		Blue:  float32(blue),
		Alpha: wrapperspb.Float(1),
	}
	return color, nil
}

func protoColorToBlock(color *colorpb.Color) []interface{} {
	if color == nil {
		return []interface{}{}
	}
	raw := map[string]interface{}{
		"red":   normalizeColorChannel(color.GetRed()),
		"green": normalizeColorChannel(color.GetGreen()),
		"blue":  normalizeColorChannel(color.GetBlue()),
	}
	return []interface{}{raw}
}

func normalizeColorChannel(value float32) float64 {
	return math.Round(float64(value)*1_000_000) / 1_000_000
}

func colorChannel(raw map[string]interface{}, key string) (float64, error) {
	value, err := floatValue(raw[key])
	if err != nil {
		return 0, errors.Errorf("invalid color %s", key)
	}
	if value < 0 || value > 1 {
		return 0, errors.Errorf("invalid color %s %v, want value between 0 and 1", key, value)
	}
	return value, nil
}

func floatValue(raw interface{}) (float64, error) {
	switch v := raw.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	default:
		return 0, errors.New("invalid float value")
	}
}

func colorBlockSchema(description string, computed bool) *schema.Schema {
	result := &schema.Schema{
		Type:        schema.TypeList,
		Optional:    !computed,
		Computed:    computed,
		Description: description,
		Elem: &schema.Resource{
			Schema: colorFieldSchema(computed),
		},
	}
	if !computed {
		result.MaxItems = 1
		result.ConfigMode = schema.SchemaConfigModeBlock
	}
	return result
}

func colorFieldSchema(computed bool) map[string]*schema.Schema {
	fields := map[string]*schema.Schema{
		"red": {
			Type:        schema.TypeFloat,
			Required:    !computed,
			Computed:    computed,
			Description: "The amount of red in the color as a value in the interval [0, 1].",
		},
		"green": {
			Type:        schema.TypeFloat,
			Required:    !computed,
			Computed:    computed,
			Description: "The amount of green in the color as a value in the interval [0, 1].",
		},
		"blue": {
			Type:        schema.TypeFloat,
			Required:    !computed,
			Computed:    computed,
			Description: "The amount of blue in the color as a value in the interval [0, 1].",
		},
	}
	if !computed {
		for _, field := range fields {
			field.ValidateFunc = validation.FloatBetween(0, 1)
		}
	}
	return fields
}
