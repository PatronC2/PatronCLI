package common

import (
	"reflect"
	"testing"
)

func TestQueryFields_SelectAndTrim(t *testing.T) {
	items := []map[string]interface{}{
		{"id": 1, "hostname": "alpha", "ip": "10.0.0.1"},
		{"id": 2, "hostname": "bravo", "ip": "10.0.0.2"},
	}

	got := QueryFields(items, "hostname, ip, missing")
	want := []map[string]interface{}{
		{"hostname": "alpha", "ip": "10.0.0.1"},
		{"hostname": "bravo", "ip": "10.0.0.2"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected query fields result\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestQueryFields_EmptyItems(t *testing.T) {
	got := QueryFields(nil, "id")
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %#v", got)
	}
}
