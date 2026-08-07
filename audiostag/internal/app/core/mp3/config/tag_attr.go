package config

import (
	"errors"
	"log/slog"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/michaeluuong/utilize/stringy"
)

const TagSeparator = "~"

type TagAttributes struct {
	AlbumFolderName        string
	CoverSource            string
	CoverArtist            string
	CoverAlbum             string
	KeepTag                string
	IsPlaylist             bool
	MultiDisc              string
	NoBrainz               bool
	NoDirectoryRename      bool
	PlaylistName           string
	PreConditions          string
	PreReplacements        string
	Replacements           string
	SingleArtist           string
	StringCase             stringy.StringCase
	Tags                   string
	TotalTracks            int
	VariousArtists         string
	tagIDToValues          map[string][]string
	oldReplacements        string
	oldValues              map[ValueType]string
	preConditionIDToValues map[string][]PreConditions
	replaceIDToValues      map[string][]Replacements
	preReplaceIDToValues   map[string][]Replacements
}

type Replacements struct {
	RegEx   string
	Replace string
}

type PreConditions struct {
	RegEx            string
	DestTag          string
	DestTagValue     string
	DestTagElseValue *string
}

type ValueType int

const (
	Tag ValueType = iota
	PreCondition
	Replace
	PreReplace
)

var modTags = NewModTags()

func (t *TagAttributes) IsOld(newValue string, valueType ValueType) bool {
	if t.oldValues == nil {
		t.oldValues = make(map[ValueType]string)

	}

	var oldValue string
	if value, ok := t.oldValues[valueType]; ok {
		oldValue = value

	}

	if newValue == oldValue {
		return true

	}

	return false

}

func (t *TagAttributes) IsKeeper(id string) bool {
	keepers := strings.Split(t.KeepTag, TagSeparator)
	keepSet := make(map[string]bool)
	for _, id := range keepers {
		keepSet[id] = true

	}

	return keepSet[id]

}

func (t *TagAttributes) SplitTag() map[string][]string {
	if !t.IsOld(t.Tags, Tag) {
		t.tagIDToValues = attrTagIDToValues(t.Tags)
		t.oldValues[Tag] = t.Tags

	}

	if t.tagIDToValues == nil {
		t.tagIDToValues = make(map[string][]string)

	}

	return t.tagIDToValues

}

func (t *TagAttributes) SplitPreConditions() (map[string][]PreConditions, error) {
	//var conditionIDToSliceValues map[string][]string
	preConditions := []string{}
	if !t.IsOld(t.PreConditions, PreCondition) {
		preConditions = splitValue(t.PreConditions)
		t.oldValues[PreCondition] = t.PreConditions

	}

	if t.preConditionIDToValues == nil {
		t.preConditionIDToValues = make(map[string][]PreConditions)

	}

	for _, preCondition := range preConditions {
		values := strings.Split(preCondition, "=")
		if len(values) > 3 {
			id := tagNameOrID(values[0])
			if id == "" {
				slog.Error("tagNameOrId()|could not find name or id", "preCondition", preCondition)
				continue

			}

			destTag := tagNameOrID(values[2])
			if destTag == "" {
				slog.Error("tagNameOrId()|could not find name or id or destination tag", "preCondition", preCondition)
				continue

			}
			if destTag == "TPOS" {
				t.MultiDisc = "true"

			}

			preConditions := PreConditions{
				RegEx:        values[1],
				DestTag:      destTag,
				DestTagValue: values[3],
			}

			if len(values) > 4 {
				preConditions.DestTagElseValue = &values[4]

			}

			t.preConditionIDToValues[id] = append(t.preConditionIDToValues[id], preConditions)

		}

	}

	return t.preConditionIDToValues, nil

}

func (t *TagAttributes) AddReplacement(ID, regex, replace string) {
	if t.replaceIDToValues == nil {
		t.replaceIDToValues = make(map[string][]Replacements)

	}

	t.replaceIDToValues[ID] = append(t.replaceIDToValues[ID], Replacements{
		RegEx:   regex,
		Replace: replace,
	})

}

func (t *TagAttributes) SplitPreReplacements(addReplacements ...string) (map[string][]Replacements, error) {
	return t.splitReplacements(true, addReplacements...)

}

func (t *TagAttributes) SplitReplacements(addReplacements ...string) (map[string][]Replacements, error) {
	return t.splitReplacements(false, addReplacements...)

}

func (t *TagAttributes) splitReplacements(isPre bool, addReplacements ...string) (map[string][]Replacements, error) {
	var replaceIDToSliceValues map[string][]string

	replacements := t.Replacements
	replaceIDToValues := &t.replaceIDToValues
	replaceType := Replace
	if isPre {
		replacements = t.PreReplacements
		replaceIDToValues = &t.preReplaceIDToValues
		replaceType = PreReplace

	}

	if !t.IsOld(replacements, replaceType) {
		replaceIDToSliceValues = attrTagIDToValues(replacements)
		t.oldValues[replaceType] = replacements
		//t.oldReplacements = t.Replacements

	} else if len(addReplacements) > 0 && !isPre {
		replaceIDToSliceValues = attrTagIDToValues(addReplacements[0])

	}

	if *replaceIDToValues == nil {
		*replaceIDToValues = make(map[string][]Replacements)

	}

	slog.Debug("attrTagIDToValues()", "replaceIDToSliceValues", replaceIDToSliceValues)
	for ID, values := range replaceIDToSliceValues {
		if len(values)%2 != 0 {
			return nil, errors.New("every replacement must be in [name=regular expression=replacement] format")

		}

		for i := 0; i < len(values); {
			(*replaceIDToValues)[ID] = append((*replaceIDToValues)[ID], Replacements{
				RegEx:   values[i],
				Replace: values[i+1],
			})

			i += 2

		}

	}
	return *replaceIDToValues, nil

}

func tagNameOrID(tagNameOrID string) string {
	titler := cases.Title(language.English)
	tagNameTitle := titler.String(tagNameOrID)

	var id string
	var ok bool
	if id, ok = modTags.IsName(tagNameTitle); !ok {
		idUpper := strings.ToUpper(tagNameOrID)
		if _, ok = modTags.IsID(idUpper); ok {
			id = idUpper

		}

	}

	if !ok {
		slog.Error("could not find name or id", "tagNameOrID", tagNameOrID)

	}

	return id

}

func splitValue(value string) []string {
	values := []string{}
	if value != "" {
		for tagValue := range strings.SplitSeq(value, TagSeparator) {
			values = append(values, tagValue)

		}

	}

	return values

}

func attrTagIDToValues(tag string) map[string][]string {
	tags := make(map[string][]string)

	if tag != "" {
		//modTags := NewModTags()
		titler := cases.Title(language.English)

		for tagValue := range strings.SplitSeq(tag, TagSeparator) {
			tagValueParts := strings.Split(tagValue, "=")
			if len(tagValueParts) >= 2 {
				//tagName, tagValue := tagValueParts[0], tagValueParts[1]
				tagName := tagValueParts[0]
				tagNameTitle := titler.String(tagName)
				var id string
				var ok bool
				if id, ok = modTags.IsName(tagNameTitle); !ok {
					idUpper := strings.ToUpper(tagName)
					if _, ok = modTags.IsID(idUpper); ok {
						id = idUpper

					}

				}

				if !ok {
					slog.Error("could not find name or id", "tagName", tagName)
					continue

				}

				var tagValues []string
				if values, ok := tags[id]; ok {
					tagValues = values

				}
				tagValues = append(tagValues, tagValueParts[1:]...)
				tags[id] = tagValues

			}

		}

	}

	return tags

}
