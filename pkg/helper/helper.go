package helper

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func IntPointerToInt64Pointer(value *int) *int64 {
	if value == nil {
		return nil
	}
	ret := int64(*value)
	return &ret
}

func Int64ToIntPointer(value int64) *int {
	if value == 0 {
		return nil
	}
	ret := int(value)
	return &ret
}

// API data types to TF types
func SetIntToInt64OrNull(value int) types.Int64 {
	if value == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(value))
}

func SliceStringToSliceTypesString(input []string) []types.String {
	result := make([]types.String, len(input))
	for i, v := range input {
		result[i] = types.StringValue(v)
	}
	return result
}

// TF types to API data types
func Int64SetToIntSlice(set types.Set) []int {
	elements := set.Elements()
	result := make([]int, len(elements))
	for i, el := range elements {
		result[i] = int(el.(types.Int64).ValueInt64())
	}
	return result
}

func StringSetToStringSlice(set types.Set) []string {
	elements := set.Elements()
	result := make([]string, len(elements))
	for i, el := range elements {
		result[i] = el.(types.String).ValueString()
	}
	return result
}

func TypesInt64ToInt64Pointer(value types.Int64) *int64 {
	if value.IsNull() {
		return nil
	}
	fieldVal := value.ValueInt64()
	return &fieldVal
}

func TypesStringSliceToStringSlice(list []types.String) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.ValueString()
	}
	return result
}

// useful for docs
func DocString(inp string) string {
	newString := strings.ReplaceAll(inp, "~~~", "`")
	return regexp.MustCompile(`(?m)^\t+`).ReplaceAllString(newString, "")
}

// returns trueVal, if condition == true, else falseVal
func TernaryOperator[T any](condition bool, trueVal T, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func SliceStringToTypesListInt64Value(slice []string) (types.List, diag.Diagnostics) {
	if len(slice) == 0 {
		return types.ListNull(types.Int64Type), diag.Diagnostics{}
	}
	attrValues := make([]attr.Value, len(slice))
	for i, v := range slice {
		val, _ := strconv.Atoi(v)
		attrValues[i] = types.Int64Value(int64(val))
	}
	return types.ListValue(types.Int64Type, attrValues)
}

func TypesListInt64SliceToInt64Slice(list types.List) []int64 {
	elements := list.Elements()
	result := make([]int64, len(elements))
	for i, v := range elements {
		result[i] = v.(types.Int64).ValueInt64()
	}
	return result
}

func TypesListStringToStringSlice(list types.List) []string {
	elements := list.Elements()
	result := make([]string, len(elements))
	for i, v := range elements {
		result[i] = v.(types.String).ValueString()
	}
	return result
}

func SliceStringToTypesListStringValue(slice []string) (types.List, diag.Diagnostics) {
	attrValues := make([]attr.Value, len(slice))
	for i, v := range slice {
		attrValues[i] = types.StringValue(v)
	}
	return types.ListValue(types.StringType, attrValues)
}

func SliceStringToSliceInt64(slice []string) []int64 {
	result := make([]int64, len(slice))
	for i, v := range slice {
		result[i], _ = strconv.ParseInt(v, 10, 64)
	}
	return result
}
