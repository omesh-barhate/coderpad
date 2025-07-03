package crdt

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Document struct {
	Characters []Character
}

type Character struct {
	ID      string
	Visible bool
	Value   string
	PrevID  string
	NextID  string
}

var (
	SiteID = 0

	LocalClock = 0

	StartCharacter = Character{ID: "start", Visible: false, Value: "", PrevID: "", NextID: "end"}

	EndCharacter = Character{ID: "end", Visible: false, Value: "", PrevID: "start", NextID: ""}

	ErrOutOfBounds    = errors.New("position out of bounds")
	ErrEmptyCharacter = errors.New("empty char ID provided")
	ErrBoundsMissing  = errors.New("subsequence bound(s) not present")
)

func New() Document {
	return Document{Characters: []Character{StartCharacter, EndCharacter}}
}

func Load(fileName string) (Document, error) {
	document := New()
	content, err := os.ReadFile(fileName)
	if err != nil {
		return document, err
	}
	lines := strings.Split(string(content), "\n")
	position := 1
	for i := 0; i < len(lines); i++ {
		for j := 0; j < len(lines[i]); j++ {
			_, err := document.Insert(position, string(lines[i][j]))
			if err != nil {
				return document, err
			}
			position++
		}
		if i < len(lines)-1 {
			_, err := document.Insert(position, "\n")
			if err != nil {
				return document, err
			}
			position++
		}
	}
	return document, err
}

func Save(fileName string, document *Document) error {
	return os.WriteFile(fileName, []byte(Content(*document)), 0644)
}

func (document *Document) SetText(newDocument Document) {
	for _, character := range newDocument.Characters {
		c := Character{ID: character.ID, Visible: character.Visible, Value: character.Value, PrevID: character.PrevID, NextID: character.NextID}
		document.Characters = append(document.Characters, c)
	}
}

func Content(document Document) string {
	var builder strings.Builder
	for _, character := range document.Characters {
		if character.Visible {
			builder.WriteString(character.Value)
		}
	}
	return builder.String()
}

func IthVisible(document Document, visiblePosition int) Character {
	visibleCount := 0
	for _, character := range document.Characters {
		if character.Visible {
			if visibleCount == visiblePosition-1 {
				return character
			}
			visibleCount++
		}
	}
	return Character{ID: "-1"}
}

func (document *Document) Length() int {
	return len(document.Characters)
}

func (document *Document) ElementAt(position int) (Character, error) {
	if position < 0 || position >= document.Length() {
		return Character{}, ErrOutOfBounds
	}
	return document.Characters[position], nil
}

func (document *Document) Position(characterID string) int {
	for index, character := range document.Characters {
		if characterID == character.ID {
			return index + 1
		}
	}
	return -1
}

func (document *Document) Left(characterID string) string {
	index := document.Position(characterID)
	if index <= 0 {
		return document.Characters[index].ID
	}
	return document.Characters[index-1].ID
}

func (document *Document) Right(characterID string) string {
	index := document.Position(characterID)
	if index >= len(document.Characters)-1 {
		return document.Characters[index-1].ID
	}
	return document.Characters[index+1].ID
}

func (document *Document) Contains(characterID string) bool {
	return document.Position(characterID) != -1
}

func (document *Document) Find(id string) Character {
	for _, character := range document.Characters {
		if character.ID == id {
			return character
		}
	}
	return Character{ID: "-1"}
}

func (document *Document) Subseq(startCharacter, endCharacter Character) ([]Character, error) {
	startIndex := document.Position(startCharacter.ID)
	endIndex := document.Position(endCharacter.ID)
	if startIndex == -1 || endIndex == -1 {
		return document.Characters, ErrBoundsMissing
	}
	if startIndex > endIndex {
		return document.Characters, ErrBoundsMissing
	}
	if startIndex == endIndex {
		return []Character{}, nil
	}
	return document.Characters[startIndex : endIndex-1], nil
}

func (document *Document) LocalInsert(character Character, position int) (*Document, error) {
	if position <= 0 || position >= document.Length() {
		return document, ErrOutOfBounds
	}
	if character.ID == "" {
		return document, ErrEmptyCharacter
	}
	document.Characters = append(document.Characters[:position],
		append([]Character{character}, document.Characters[position:]...)...,
	)
	document.Characters[position-1].NextID = character.ID
	document.Characters[position+1].PrevID = character.ID
	return document, nil
}

func (document *Document) IntegrateInsert(character, prevCharacter, nextCharacter Character) (*Document, error) {
	subsequence, err := document.Subseq(prevCharacter, nextCharacter)
	if err != nil {
		return document, err
	}
	insertPosition := document.Position(nextCharacter.ID)
	insertPosition--
	if len(subsequence) == 0 {
		return document.LocalInsert(character, insertPosition)
	}
	if len(subsequence) == 1 {
		return document.LocalInsert(character, insertPosition-1)
	}
	index := 1
	for index < len(subsequence)-1 && subsequence[index].ID < character.ID {
		index++
	}
	return document.IntegrateInsert(character, subsequence[index-1], subsequence[index])
}

func (document *Document) GenerateInsert(position int, value string) (*Document, error) {
	LocalClock++
	prevCharacter := IthVisible(*document, position-1)
	nextCharacter := IthVisible(*document, position)
	if prevCharacter.ID == "-1" {
		prevCharacter = document.Find("start")
	}
	if nextCharacter.ID == "-1" {
		nextCharacter = document.Find("end")
	}
	character := Character{
		ID:      fmt.Sprint(SiteID) + fmt.Sprint(LocalClock),
		Visible: true,
		Value:   value,
		PrevID:  prevCharacter.ID,
		NextID:  nextCharacter.ID,
	}
	return document.IntegrateInsert(character, prevCharacter, nextCharacter)
}

func (document *Document) IntegrateDelete(character Character) *Document {
	position := document.Position(character.ID)
	if position == -1 {
		return document
	}
	document.Characters[position-1].Visible = false
	return document
}

func (document *Document) GenerateDelete(position int) *Document {
	character := IthVisible(*document, position)
	return document.IntegrateDelete(character)
}

func (document *Document) Insert(position int, value string) (string, error) {
	newDocument, err := document.GenerateInsert(position, value)
	if err != nil {
		return Content(*document), err
	}
	return Content(*newDocument), nil
}

func (document *Document) Delete(position int) string {
	newDocument := document.GenerateDelete(position)
	return Content(*newDocument)
}
