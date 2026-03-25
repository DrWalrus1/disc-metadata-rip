package bdmv

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// IndexBDMV represents the parsed contents of an index.bdmv file.
type IndexBDMV struct {
	Version       string
	AppInfoOffset uint32
	ExtDataOffset uint32
	FirstPlay     TitleEntry
	TopMenu       TitleEntry
	Titles        []TitleEntry
}

// TitleEntry represents a single title in the index.
type TitleEntry struct {
	ObjectType  uint8
	AccessType  uint8
	ObjectIDRef uint16
}

// IsHDMV returns true if the title uses HDMV navigation.
func (t TitleEntry) IsHDMV() bool {
	return t.ObjectType&ObjectTypeMask == ObjectTypeHDMV ||
		t.ObjectType == ObjectTypeHDMVFirstPlay
}

// IsBDJ returns true if the title uses BD-J (Java) navigation.
func (t TitleEntry) IsBDJ() bool {
	return t.ObjectType&ObjectTypeMask == ObjectTypeBDJ
}

// PlaylistPath returns the full path to the playlist file for this title.
func (t TitleEntry) PlaylistPath(bdmvRoot string) string {
	if t.IsHDMV() {
		return filepath.Join(bdmvRoot, "PLAYLIST", fmt.Sprintf("%05d.mpls", t.ObjectIDRef))
	}
	return ""
}

// ParseIndex parses the index.bdmv file at the given path.
func ParseIndex(path string) (*IndexBDMV, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	typeInd := make([]byte, 4)
	version := make([]byte, 4)
	io.ReadFull(f, typeInd)
	io.ReadFull(f, version)

	if string(typeInd) != TypeIndicatorINDX {
		return nil, fmt.Errorf("not an index.bdmv file (got %q)", typeInd)
	}

	idx := &IndexBDMV{Version: string(version)}
	binary.Read(f, binary.BigEndian, &idx.AppInfoOffset)
	binary.Read(f, binary.BigEndian, &idx.ExtDataOffset)

	f.Seek(int64(idx.AppInfoOffset)+indexAppInfoBodyOffset, io.SeekStart)

	idx.FirstPlay, err = readShortEntry(f)
	if err != nil {
		return nil, fmt.Errorf("reading FirstPlay: %w", err)
	}
	idx.TopMenu, err = readShortEntry(f)
	if err != nil {
		return nil, fmt.Errorf("reading TopMenu: %w", err)
	}

	var numTitles uint16
	binary.Read(f, binary.BigEndian, &numTitles)

	idx.Titles = make([]TitleEntry, numTitles)
	for i := range idx.Titles {
		idx.Titles[i], err = readTitleEntry(f)
		if err != nil {
			return nil, fmt.Errorf("reading title %d: %w", i, err)
		}
	}

	return idx, nil
}

func readShortEntry(r io.Reader) (TitleEntry, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return TitleEntry{}, err
	}
	return TitleEntry{
		ObjectType:  buf[0],
		AccessType:  buf[1],
		ObjectIDRef: binary.BigEndian.Uint16(buf[2:4]),
	}, nil
}

func readTitleEntry(r io.Reader) (TitleEntry, error) {
	buf := make([]byte, navCommandSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return TitleEntry{}, err
	}

	entry := TitleEntry{
		ObjectType: buf[0],
		AccessType: buf[1],
	}

	switch buf[0] & ObjectTypeMask {
	case ObjectTypeHDMV:
		refStr := string(buf[6:11])
		var val uint16
		fmt.Sscanf(refStr, "%d", &val)
		entry.ObjectIDRef = val
	case ObjectTypeBDJ:
		entry.ObjectIDRef = binary.BigEndian.Uint16(buf[6:8])
	}

	return entry, nil
}
