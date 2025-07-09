package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReadRequestJSON(t *testing.T) {
	req := ReadRequest{Cid: "c1", Item: "x"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var decoded ReadRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(req, decoded) {
		t.Errorf("Expected %+v, got %+v", req, decoded)
	}
}

func TestCommitDecisionJSON(t *testing.T) {
	cd := CommitDecision{Cid: "c1", Tid: "t1", Commit: true}
	data, err := json.Marshal(cd)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var decoded CommitDecision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(cd, decoded) {
		t.Errorf("Expected %+v, got %+v", cd, decoded)
	}
}
