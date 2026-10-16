package webhook

import (
	"strings"
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
func TestPreserveOrderKeepsConfiguredOrder(t *testing.T) {
	desired := jobIDs(30, 10, 20)
	fromAPI := jobIDs(10, 20, 30)

	if got := ids(preserveOrder(desired, fromAPI)); !equal(got, []int64{30, 10, 20}) {
		t.Errorf("got %v, want the configured order [30 10 20]", got)
	}
}

// A real change outside Terraform still has to surface.
func TestPreserveOrderTakesApiWhenIdsDiffer(t *testing.T) {
	desired := jobIDs(30, 10, 20)
	fromAPI := jobIDs(10, 20, 40)

	if got := ids(preserveOrder(desired, fromAPI)); !equal(got, []int64{10, 20, 40}) {
		t.Errorf("got %v, want the API ids [10 20 40]", got)
	}
}

func TestPreserveOrderTakesApiWhenCountsDiffer(t *testing.T) {
	desired := jobIDs(10, 20)
	fromAPI := jobIDs(10, 20, 30)

	if got := ids(preserveOrder(desired, fromAPI)); !equal(got, []int64{10, 20, 30}) {
		t.Errorf("got %v, want the API ids [10 20 30]", got)
	}
}

// Import and the first read have nothing to preserve.
func TestPreserveOrderWithoutPriorValue(t *testing.T) {
	fromAPI := jobIDs(10, 20)

	for name, desired := range map[string]types.List{
		"null":    types.ListNull(types.Int64Type),
		"unknown": types.ListUnknown(types.Int64Type),
	} {
		if got := ids(preserveOrder(desired, fromAPI)); !equal(got, []int64{10, 20}) {
			t.Errorf("%s: got %v, want the API ids [10 20]", name, got)
		}
	}
}

// A repeated id must not make two different lists look the same.
func TestSameElementsCountsDuplicates(t *testing.T) {
	if sameElements(jobIDs(10, 10, 20), jobIDs(10, 20, 20)) {
		t.Error("lists with different duplicate counts were treated as equal")
	}
}

func eventTypes(values ...string) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elements)
}

func names(list types.List) []string {
	out := make([]string, 0, len(list.Elements()))
	for _, element := range list.Elements() {
		out = append(out, element.(types.String).ValueString())
	}
	return out
}

// event_types come back reordered the same way as job_ids.
func TestPreserveOrderKeepsConfiguredEventTypeOrder(t *testing.T) {
	desired := eventTypes("job.run.completed", "job.run.started", "job.run.errored")
	fromAPI := eventTypes("job.run.started", "job.run.errored", "job.run.completed")

	got := names(preserveOrder(desired, fromAPI))
	if strings.Join(got, ",") != "job.run.completed,job.run.started,job.run.errored" {
		t.Errorf("got %v, want the configured order", got)
	}
}

func TestPreserveOrderTakesApiWhenEventTypesDiffer(t *testing.T) {
	desired := eventTypes("job.run.completed", "job.run.started")
	fromAPI := eventTypes("job.run.started", "job.run.errored")

	got := names(preserveOrder(desired, fromAPI))
	if strings.Join(got, ",") != "job.run.started,job.run.errored" {
		t.Errorf("got %v, want the API event types", got)
	}
}
