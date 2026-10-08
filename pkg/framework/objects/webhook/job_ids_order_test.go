package webhook

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func jobIDs(values ...int64) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.Int64Value(v))
	}
	return types.ListValueMust(types.Int64Type, elements)
}

func ids(list types.List) []int64 {
	out := make([]int64, 0, len(list.Elements()))
	for _, element := range list.Elements() {
		out = append(out, element.(types.Int64).ValueInt64())
	}
	return out
}

func equal(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// The API returns the same ids in another order, which must not read as a change.
func TestPreserveJobIDOrderKeepsConfiguredOrder(t *testing.T) {
	desired := jobIDs(30, 10, 20)
	fromAPI := jobIDs(10, 20, 30)

	if got := ids(preserveJobIDOrder(desired, fromAPI)); !equal(got, []int64{30, 10, 20}) {
		t.Errorf("got %v, want the configured order [30 10 20]", got)
	}
}

// A real change outside Terraform still has to surface.
func TestPreserveJobIDOrderTakesApiWhenIdsDiffer(t *testing.T) {
	desired := jobIDs(30, 10, 20)
	fromAPI := jobIDs(10, 20, 40)

	if got := ids(preserveJobIDOrder(desired, fromAPI)); !equal(got, []int64{10, 20, 40}) {
		t.Errorf("got %v, want the API ids [10 20 40]", got)
	}
}

func TestPreserveJobIDOrderTakesApiWhenCountsDiffer(t *testing.T) {
	desired := jobIDs(10, 20)
	fromAPI := jobIDs(10, 20, 30)

	if got := ids(preserveJobIDOrder(desired, fromAPI)); !equal(got, []int64{10, 20, 30}) {
		t.Errorf("got %v, want the API ids [10 20 30]", got)
	}
}

// Import and the first read have nothing to preserve.
func TestPreserveJobIDOrderWithoutPriorValue(t *testing.T) {
	fromAPI := jobIDs(10, 20)

	for name, desired := range map[string]types.List{
		"null":    types.ListNull(types.Int64Type),
		"unknown": types.ListUnknown(types.Int64Type),
	} {
		if got := ids(preserveJobIDOrder(desired, fromAPI)); !equal(got, []int64{10, 20}) {
			t.Errorf("%s: got %v, want the API ids [10 20]", name, got)
		}
	}
}

// A repeated id must not make two different lists look the same.
func TestSameJobIDsCountsDuplicates(t *testing.T) {
	if sameJobIDs(jobIDs(10, 10, 20), jobIDs(10, 20, 20)) {
		t.Error("lists with different duplicate counts were treated as equal")
	}
}
