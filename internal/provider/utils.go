package provider

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func convertTFStringsToGoStrings(tfStrings []types.String) []string {
	goStrings := make([]string, 0, len(tfStrings))
	for _, tfStr := range tfStrings {
		if !tfStr.IsNull() && !tfStr.IsUnknown() {
			goStrings = append(goStrings, tfStr.ValueString())
		}
	}
	return goStrings
}
