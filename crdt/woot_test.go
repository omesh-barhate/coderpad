package crdt

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocument(t *testing.T) {
	document := New()
	length := document.Length()
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("length mismatch; got = %v, expected = %v\n", length, expectedLength)
	}
}

func TestInsert(t *testing.T) {
	document := New()
	position := 1
	value := "a"
	content, err := document.Insert(position, value)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	expectedDocument := &Document{
		Characters: []Character{
			{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "end"},
			{ID: "1", Visible: true, Value: "a", PrevID: "start", NextID: "end"},
			{ID: "end", Visible: false, Value: "", PrevID: "1", NextID: ""},
		},
	}
	got := content
	want := Content(*expectedDocument)
	if got != want {
		t.Errorf("content mismatch; got = %v, expected = %v\n", got, want)
	}
}

func TestIntegrateInsert_SamePosition(t *testing.T) {
	document := &Document{
		Characters: []Character{
			{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "1"},
			{ID: "1", Visible: false, Value: "e", PrevID: "start", NextID: "2"},
			{ID: "2", Visible: false, Value: "n", PrevID: "1", NextID: "end"},
			{ID: "end", Visible: false, Value: "", PrevID: "2", NextID: ""},
		},
	}
	newCharacter := Character{ID: "3", Visible: false, Value: "b", PrevID: "start", NextID: "1"}
	prevCharacter := Character{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "1"}
	nextCharacter := Character{ID: "1", Visible: false, Value: "e", PrevID: "start", NextID: "2"}
	resultDocument, err := document.IntegrateInsert(newCharacter, prevCharacter, nextCharacter)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	expectedDocument := &Document{
		Characters: []Character{
			{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "3"},
			{ID: "3", Visible: false, Value: "b", PrevID: "start", NextID: "1"},
			{ID: "1", Visible: false, Value: "e", PrevID: "3", NextID: "2"},
			{ID: "2", Visible: false, Value: "n", PrevID: "1", NextID: "end"},
			{ID: "end", Visible: false, Value: "", PrevID: "2", NextID: ""},
		},
	}
	if !cmp.Equal(resultDocument, expectedDocument) {
		t.Errorf("document mismatch; diff = %v\n", cmp.Diff(resultDocument, expectedDocument))
	}
}

func TestIntegrateInsert_BetweenTwoPositions(t *testing.T) {
	document := &Document{
		Characters: []Character{
			{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "1"},
			{ID: "1", Visible: false, Value: "c", PrevID: "start", NextID: "2"},
			{ID: "2", Visible: false, Value: "t", PrevID: "1", NextID: "end"},
			{ID: "end", Visible: false, Value: "", PrevID: "2", NextID: ""},
		},
	}
	newCharacter := Character{ID: "3", Visible: false, Value: "a", PrevID: "1", NextID: "2"}
	prevCharacter := Character{ID: "1", Visible: false, Value: "c", PrevID: "start", NextID: "2"}
	nextCharacter := Character{ID: "2", Visible: false, Value: "t", PrevID: "1", NextID: "end"}
	resultDocument, err := document.IntegrateInsert(newCharacter, prevCharacter, nextCharacter)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	expectedDocument := &Document{
		Characters: []Character{
			{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "1"},
			{ID: "1", Visible: false, Value: "c", PrevID: "start", NextID: "3"},
			{ID: "3", Visible: false, Value: "a", PrevID: "1", NextID: "2"},
			{ID: "2", Visible: false, Value: "t", PrevID: "3", NextID: "end"},
			{ID: "end", Visible: false, Value: "", PrevID: "2", NextID: ""},
		},
	}
	if !cmp.Equal(resultDocument, expectedDocument) {
		t.Errorf("document mismatch; diff = %v\n", cmp.Diff(resultDocument, expectedDocument))
	}
}

func TestLoad(t *testing.T) {
	document := &Document{
		Characters: []Character{
			{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "1"},
			{ID: "1", Visible: true, Value: "c", PrevID: "start", NextID: "3"},
			{ID: "3", Visible: true, Value: "a", PrevID: "1", NextID: "2"},
			{ID: "2", Visible: true, Value: "t", PrevID: "3", NextID: "4"},
			{ID: "4", Visible: true, Value: "\n", PrevID: "2", NextID: "5"},
			{ID: "5", Visible: true, Value: "d", PrevID: "4", NextID: "6"},
			{ID: "6", Visible: true, Value: "o", PrevID: "5", NextID: "7"},
			{ID: "7", Visible: true, Value: "g", PrevID: "6", NextID: "end"},
			{ID: "end", Visible: false, Value: "", PrevID: "7", NextID: ""},
		},
	}
	tmpFile, err := os.CreateTemp("", "ex")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	defer os.Remove(tmpFile.Name())
	err = Save(tmpFile.Name(), document)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}
	loadedDocument, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}
	got := Content(loadedDocument)
	want := Content(*document)
	if !cmp.Equal(got, want) {
		t.Errorf("content mismatch; diff = %v\n", cmp.Diff(got, want))
	}
}
