package helper

import (
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EmptySetDefault(elemType attr.Type) defaults.Set {
	return setdefault.StaticValue(
		types.SetValueMust(
			elemType,
			[]attr.Value{},
		),
	)
}

func SetIntToInt64OrNull(value int) types.Int64 {
	if value == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(value))
}

func Int64SetToIntSlice(set types.Set) []int {
	elements := set.Elements()
	result := make([]int, len(elements))
	for i, el := range elements {
		result[i] = int(el.(types.Int64).ValueInt64())
	}
	return result
}

func DocString(inp string) string {
	newString := strings.ReplaceAll(inp, "~~~", "`")
	return regexp.MustCompile(`(?m)^\t+`).ReplaceAllString(newString, "")
}

func IntPointerToInt64Pointer(value *int) *int64 {
	if value == nil {
		return nil
	}
	ret := int64(*value)
	return &ret
}
