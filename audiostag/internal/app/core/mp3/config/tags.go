package config

import (
	"reflect"

	"github.com/cabbagekobe/tunetag/id3v2"

	"github.com/michaeluuong/utilize/reflections"
)

type ReqTags struct {
	IDToName     map[string]string   `json:"required_tags"`
	CopyToFields map[string][]string `json:"copy_to_fields"`
	IdV2ToIDV1   map[string]string
}

// IsID deterimines if an ID is a required tag.
//   - id is the 4 character tag ID (e.g. TRCK, APIC, TIT2, etc.)
//
// Return true if the id is a required tag or false if it is not.
func (b ReqTags) IsID(id string) bool {
	_, ok := b.IDToName[id]
	return ok

}

// Is2In1 determines if an ID3V2 tag ID has a corresponding ID3V1 tag.
//   - idV2 is the 4 character ID3V2 tag ID
//
// Return
//   - IDV1 ID if there is a corresponding tag
//   - true if the ID3V2 tag has a corresponding ID3V1 tag or false if it doesn't
func (b ReqTags) Is2In1(idV2 string) (string, bool) {
	V1ID, ok := b.IdV2ToIDV1[idV2]
	return V1ID, ok

}

// IsCopyTo determines if
func (b ReqTags) IsCopyTo(key string) bool {
	_, ok := b.CopyToFields[key]
	return ok

}

var RequiredTags = &ReqTags{
	IDToName: map[string]string{
		"TIT2": "Title",
		"TPE1": "Artist",
		"TALB": "Album",
		"TOAL": "Original Album",
		"COMM": "Comment",
		"TYER": "Year",
		"TDRC": "Recording Time",
		"TRCK": "Track Number",
		"TCON": "Genre",
		"TPE2": "Album Artist",
		"TOPE": "Original Artist",
		"TPOS": "Disc Number",
		"APIC": "Picture",
		//"TPUB": "Publisher",
		"TCOM": "Composer",
	},

	IdV2ToIDV1: map[string]string{
		"TIT2": "Title",
		"TPE1": "Artist",
		"TALB": "Album",
		"COMM": "Comment",
		"TYER": "Year",
		"TDRC": "Year",
		"TRCK": "Track",
		"TCON": "Genre",
	},

	CopyToFields: map[string][]string{
		"TPE1": {"TPE2"},
	},
}

type Tags struct {
	NameToID map[string]string
	IDToName map[string]string
}

func (t Tags) IsName(name string) (string, bool) {
	id, ok := t.NameToID[name]
	return id, ok

}

func (t Tags) Name(id string) string {
	name, ok := t.IsID(id)
	if !ok {
		name = ""

	}

	return name

}

func (t Tags) IsID(id string) (string, bool) {
	name, ok := t.IDToName[id]
	return name, ok

}

var NameToIDMod = map[string]string{
	"Arranger":                          "IPLS", // Kid3
	"Author":                            "TOLY", // Kid3
	"Chapter":                           "CHAP", // Kid3
	"Table Of Contents":                 "CTOC", // Kid3
	"Compilation":                       "TCMP", // Kid3
	"Description":                       "TIT3", // Kid3
	"Commercial Url":                    "WCOM", // Kid3
	"File":                              "File", // audios
	"Grouping":                          "GRP1", // Kid3
	"Movement Name":                     "MVNM", // Kid3
	"Movement Number":                   "MVIN", // Kid3
	"Original Date":                     "TORY", // Kid3
	"Performer":                         "IPLS", // Kid3
	"Podcast":                           "PCST", // Kid3
	"Podcast Category":                  "TCAT", // Kid3
	"Podcast Identifier":                "TGID", // Kid3
	"Podcast Keyword":                   "TKWD", // Kid3
	"Podcast Feed":                      "WFED", // Kid3
	"Podcast Description":               "TDES", // Kid3
	"Remixer":                           "TPE4", // Kid3
	"Sort Album Artist":                 "TSO2", // Kid3
	"Sort Composer":                     "TSOC", // Kid3
	"Date":                              "TDAT",
	"User":                              "TXXX",
	"Audio Encryption":                  "AENC",
	"Picture":                           "APIC",
	"Audio Seek Point Index":            "ASPI",
	"Comment":                           "COMM",
	"Commercial Frame":                  "COMR",
	"Encryption Method Registration":    "ENCR",
	"Equalisation 2":                    "EQU2",
	"Year":                              "TYER",
	"Event Timing Codes":                "ETCO",
	"General Encapsulated Object":       "GEOB",
	"Group Identification Registration": "GRID",
	"Linked Information":                "LINK",
	"Music Cd Identifier":               "MCDI",
	"MPEG Location Lookup Table":        "MLLT",
	"Ownership Frame":                   "OWNE",
	"Private Frame":                     "PRIV",
	"Play Counter":                      "PCNT",
	"Popularimeter":                     "POPM",
	"Position Synchronisation Frame":    "POSS",
	"Recommended Buffer Size":           "RBUF",
	"Relative Volume Adjustment 2":      "RVA2",
	"Reverb":                            "RVRB",
	"Seek Frame":                        "SEEK",
	"Signature Frame":                   "SIGN",
	"Synchronised Lyrics":               "SYLT",
	"Synchronised Tempo Codes":          "SYTC",
	"Album":                             "TALB",
	"Bpm":                               "TBPM",
	"Composer":                          "TCOM",
	"Genre":                             "TCON",
	"Copyright Message":                 "TCOP",
	"Encoding Time":                     "TDEN",
	"Playlist Delay":                    "TDLY",
	"Original Release Time":             "TDOR",
	"Recording Time":                    "TDRC",
	"Release Time":                      "TDRL",
	"Tagging Time":                      "TDTG",
	"Encoded By":                        "TENC",
	"Lyricist Writer":                   "TEXT",
	"File Type":                         "TFLT",
	"Involved People List":              "TIPL",
	"Content Group Description":         "TIT1",
	"Title":                             "TIT2",
	"Subtitle Refinement":               "TIT3",
	"Initial Key":                       "TKEY",
	"Language":                          "TLAN",
	"Length":                            "TLEN",
	"Musician Credits List":             "TMCL",
	"Media Type":                        "TMED",
	"Mood":                              "TMOO",
	"Original Album":                    "TOAL",
	"Original Filename":                 "TOFN",
	"Original Lyricist":                 "TOLY",
	"Original Artist":                   "TOPE",
	"File Owner":                        "TOWN",
	"Artist":                            "TPE1",
	"Album Artist":                      "TPE2",
	"Conductor Refinement":              "TPE3",
	"Interpreted":                       "TPE4",
	"Disc Number":                       "TPOS",
	"Produced notice":                   "TPRO",
	"Publisher":                         "TPUB",
	"Track Number":                      "TRCK",
	"Internet Radio Station Name":       "TRSN",
	"Internet Radio Station Owner":      "TRSO",
	"Album Sort Order":                  "TSOA",
	"Performer Sort Order":              "TSOP",
	"Title Sort Order":                  "TSOT",
	"Isrc":                              "TSRC",
	"Software Settings Encoding":        "TSSE",
	"Unique File Identifier":            "UFID",
	"Terms Of Use":                      "USER",
	"Unsynchronised Lyrics":             "USLT",
	"Commercial Information":            "WCOM",
	"Copyright Information":             "WCOP",
	"Official Audio File Webpage":       "WOAF",
	"Official Artist Webpage":           "WOAR",
	"Official Audio Source Webpage":     "WOAS",
	"Official Internet Radio Station Homepage": "WORS",
	"Payment":                     "WPAY",
	"Publishers Official Webpage": "WPUB",
	"User Defined URL Link Frame": "WXXX",
}

var IDToNameMod = SwapKeyValue(NameToIDMod)

func SwapKeyValue[K comparable, V comparable](keyToValue map[K]V) map[V]K {
	valueToKey := make(map[V]K)
	for k, v := range keyToValue {
		valueToKey[v] = k

	}

	return valueToKey

}

// NewModTags holds edited tag descriptions per ID.
func NewModTags() *Tags {
	/*idToName := make(map[string]string)
	for k, v := range nameToIDMod {
		idToName[v] = k

	}*/
	//idToNameMod := SwapKeyValue(nameToIDMod)

	return &Tags{
		NameToID: NameToIDMod,
		IDToName: IDToNameMod,
	}

}

func NewOfficialTags() *Tags {
	IdToName := map[string]string{
		"AENC": "Audio encryption",
		"APIC": "Attached picture",
		"ASPI": "Audio seek point index",
		"COMM": "Comments",
		"COMR": "Commercial frame",
		"ENCR": "Encryption method registration",
		"EQU2": "Equalisation (2)",
		"ETCO": "Event timing codes",
		"GEOB": "General encapsulated object",
		"GRID": "Group identification registration",
		"LINK": "Linked information",
		"MCDI": "Music CD identifier",
		"MLLT": "MPEG location lookup table",
		"OWNE": "Ownership frame",
		"PRIV": "Private frame",
		"PCNT": "Play counter",
		"POPM": "Popularimeter",
		"POSS": "Position synchronisation frame",
		"RBUF": "Recommended buffer size",
		"RVA2": "Relative volume adjustment (2)",
		"RVRB": "Reverb",
		"SEEK": "Seek frame",
		"SIGN": "Signature frame",
		"SYLT": "Synchronised lyric/text",
		"SYTC": "Synchronised tempo codes",
		"TALB": "Album/Movie/Show title",
		"TBPM": "BPM (beats per minute)",
		"TCOM": "Composer",
		"TCON": "Content type",
		"TCOP": "Copyright message",
		"TDAT": "Date",
		"TDEN": "Encoding time",
		"TDLY": "Playlist delay",
		"TDOR": "Original release time",
		"TDRC": "Recording time",
		"TDRL": "Release time",
		"TDTG": "Tagging time",
		"TENC": "Encoded by",
		"TEXT": "Lyricist/Text writer",
		"TFLT": "File type",
		"TIPL": "Involved people list",
		"TIT1": "Content group description",
		"TIT2": "Title/songname/content description",
		"TIT3": "Subtitle/Description refinement",
		"TKEY": "Initial key",
		"TLAN": "Language(s)",
		"TLEN": "Length",
		"TMCL": "Musician credits list",
		"TMED": "Media type",
		"TMOO": "Mood",
		"TOAL": "Original album/movie/show title",
		"TOFN": "Original filename",
		"TOLY": "Original lyricist(s)/text writer(s)",
		"TOPE": "Original artist(s)/performer(s)",
		"TOWN": "File owner/licensee",
		"TPE1": "Lead performer(s)/Soloist(s)",
		"TPE2": "Band/orchestra/accompaniment",
		"TPE3": "Conductor/performer refinement",
		"TPE4": "Interpreted, remixed, or otherwise modified by",
		"TPOS": "Part of a set",
		"TPRO": "Produced notice",
		"TPUB": "Publisher",
		"TRCK": "Track number/Position in set",
		"TRSN": "Internet radio station name",
		"TRSO": "Internet radio station owner",
		"TSOA": "Album sort order",
		"TSOP": "Performer sort order",
		"TSOT": "Title sort order",
		"TSRC": "ISRC (international standard recording code)",
		"TSSE": "Encoder Settings",
		"UFID": "Unique file identifier",
		"USER": "Terms of use",
		"USLT": "Unsynchronised lyric/text transcription",
		"WCOM": "Commercial information",
		"WCOP": "Copyright/Legal information",
		"WOAF": "Official audio file webpage",
		"WOAR": "Official artist/performer webpage",
		"WOAS": "Official audio source webpage",
		"WORS": "Official Internet radio station homepage",
		"WPAY": "Payment",
		"WPUB": "Publishers official webpage",
		"WXXX": "User defined URL link frame",
	}

	/*nameToID := make(map[string]string)
	for k, v := range idToName {
		nameToID[v] = k

	}*/
	NameToID := SwapKeyValue(IdToName)

	return &Tags{
		IDToName: IdToName,
		NameToID: NameToID,
	}

}

type TagType struct {
	IDToType map[string]reflect.Type
}

func (t TagType) IDHasType(id string) (reflect.Type, bool) {
	tagType, ok := t.IDToType[id]
	return tagType, ok

}

func (t TagType) NameHasType(name string) (reflect.Type, bool) {
	if id, ok := NameToIDMod[name]; ok {
		return t.IDHasType(id)

	}

	return nil, false

}

func NewAltTagType() *TagType {
	return &TagType{
		IDToType: map[string]reflect.Type{
			"APIC": reflections.AnyType((*id3v2.PictureFrame)(nil)),
			"IPLS": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"GRP1": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"MVIN": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"MVNM": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"OWNE": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"PCST": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"POPM": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"SYLT": reflections.AnyType((*id3v2.GenericFrame)(nil)),
			"COMM": reflections.AnyType((*id3v2.CommentFrame)(nil)),
			"USLT": reflections.AnyType((*id3v2.UnsynchronisedLyricsFrame)(nil)),
			"PRIV": reflections.AnyType((*id3v2.PrivFrame)(nil)),
			"UFID": reflections.AnyType((*id3v2.UFIDFrame)(nil)),
			"WCOM": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WCOP": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WFED": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WOAF": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WOAR": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WORS": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WPAY": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"WPUB": reflections.AnyType((*id3v2.URLFrame)(nil)),
			"TXXX": reflections.AnyType((*id3v2.UserTextFrame)(nil)),
			"WXXX": reflections.AnyType((*id3v2.UserURLFrame)(nil)),
		},
	}

}
