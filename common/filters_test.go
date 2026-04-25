package common

import (
	"reflect"
	"testing"
)

func TestSplitFilter_RespectsBrackets(t *testing.T) {
	input := "status=Online,tags.env=[prod,stage],hostname=alpha"
	got := splitFilter(input, ',')
	want := []string{"status=Online", "tags.env=[prod,stage]", "hostname=alpha"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected split result\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestParseFilter_SingleAndListValues(t *testing.T) {
	got := parseFilter("status=Online,tags.env=[prod, stage]")
	want := map[string][]string{
		"status":   {"Online"},
		"tags.env": {"prod", "stage"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected parsed filter\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b"}, "b") {
		t.Fatal("expected contains to return true")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Fatal("expected contains to return false")
	}
}

func TestFilterItemsWithTags_EmptyFilterReturnsAll(t *testing.T) {
	items := []map[string]interface{}{
		{"id": "1"},
		{"id": "2"},
	}
	got := FilterItemsWithTags(items, "")
	if !reflect.DeepEqual(got, items) {
		t.Fatalf("expected all items, got %#v", got)
	}
}

func TestFilterItemsWithTags_FieldFilter(t *testing.T) {
	items := []map[string]interface{}{
		{"id": "1", "status": "Online"},
		{"id": "2", "status": "Offline"},
	}

	got := FilterItemsWithTags(items, "status=Online")
	if len(got) != 1 || got[0]["id"] != "1" {
		t.Fatalf("unexpected filter result: %#v", got)
	}
}

func TestFilterItemsWithTags_TagFilterWithList(t *testing.T) {
	items := []map[string]interface{}{
		{
			"id": "1",
			"tags": []interface{}{
				map[string]interface{}{"key": "env", "value": "prod"},
				map[string]interface{}{"key": "team", "value": "red"},
			},
		},
		{
			"id": "2",
			"tags": []interface{}{
				map[string]interface{}{"key": "env", "value": "dev"},
			},
		},
	}

	got := FilterItemsWithTags(items, "tags.env=[prod,stage]")
	if len(got) != 1 || got[0]["id"] != "1" {
		t.Fatalf("unexpected tag filter result: %#v", got)
	}
}

func TestFilterItemsWithTags_InvalidSyntaxReturnsNil(t *testing.T) {
	items := []map[string]interface{}{{"id": "1"}}
	got := FilterItemsWithTags(items, "this-is-not-valid")
	if got != nil {
		t.Fatalf("expected nil for invalid filter, got %#v", got)
	}
}
