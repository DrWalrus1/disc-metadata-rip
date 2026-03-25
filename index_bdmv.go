package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type IndexBDMV struct {
	Version       string
	AppInfoOffset uint32
	ExtDataOffset uint32
	FirstPlay     TitleEntry
	TopMenu       TitleEntry
	Titles        []TitleEntry
}

type TitleEntry struct {
	ObjectType  uint8
	AccessType  uint8
	ObjectIDRef uint16
}

func (t TitleEntry) IsHDMV() bool {
	return t.ObjectType&objectTypeMask == objectTypeHDMV ||
		t.ObjectType == objectTypeHDMVFirstPlay
}

func (t TitleEntry) IsBDJ() bool {
	return t.ObjectType&objectTypeMask == objectTypeBDJ
}

func (t TitleEntry) PlaylistPath(bdmvRoot string) string {
	if t.IsHDMV() {
		return filepath.Join(bdmvRoot, "PLAYLIST", fmt.Sprintf("%05d.mpls", t.ObjectIDRef))
	}
	return ""
}

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

	if string(typeInd) != typeIndicatorINDX {
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

	switch buf[0] & objectTypeMask {
	case objectTypeHDMV:
		refStr := string(buf[6:11])
		var val uint16
		fmt.Sscanf(refStr, "%d", &val)
		entry.ObjectIDRef = val
	case objectTypeBDJ:
		entry.ObjectIDRef = binary.BigEndian.Uint16(buf[6:8])
	}

	return entry, nil
}
