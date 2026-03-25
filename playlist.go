package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Playlist struct {
	Name      string
	PlayItems []PlayItem
	Marks     []PlaylistMark
}

type PlayItem struct {
	ClipName string
	InTime   uint32
	OutTime  uint32
	Duration int // seconds
}

type PlaylistMark struct {
	MarkType    uint8
	PlayItemRef uint16
	Timestamp   uint32
	Duration    uint32
}

// TotalDuration returns the sum of all PlayItem durations.
func (p *Playlist) TotalDuration() int {
	total := 0
	for _, item := range p.PlayItems {
		total += item.Duration
	}
	return total
}

// EpisodeDuration returns the total duration of unique clips,
// skipping duplicates and short bumper/tail clips under 60 seconds.
func (p *Playlist) EpisodeDuration() int {
	seen := map[string]bool{}
	total := 0
	for _, item := range p.PlayItems {
		if seen[item.ClipName] || item.Duration < 60 {
			continue
		}
		seen[item.ClipName] = true
		total += item.Duration
	}
	return total
}

// PrimaryClip returns the first clip with a duration over 60 seconds.
func (p *Playlist) PrimaryClip() string {
	for _, item := range p.PlayItems {
		if item.Duration >= 60 {
			return item.ClipName
		}
	}
	if len(p.PlayItems) > 0 {
		return p.PlayItems[0].ClipName
	}
	return ""
}

// EstimateDuration estimates total stream duration from file size.
func (p *Playlist) EstimateDuration(bdmvRoot string) int {
	clip := p.PrimaryClip()
	if clip == "" {
		return 0
	}
	path := filepath.Join(bdmvRoot, "STREAM", clip+".m2ts")
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(info.Size() * 8 / estimatedBitrate)
}

// EstimateEpisodeCount checks if the stream duration is a near-integer
// multiple of the cluster episode duration, indicating a combined stream.
func (p *Playlist) EstimateEpisodeCount(bdmvRoot string, clusterDur int) int {
	if clusterDur <= 0 {
		return 1
	}

	totalDur := p.EstimateDuration(bdmvRoot)
	if totalDur < minViableDuration {
		return 1
	}

	ratio := float64(totalDur) / float64(clusterDur)
	rounded := int(ratio + 0.5)

	if rounded < 1 || rounded > maxEpisodesPerStream {
		return 1
	}

	// Relative tolerance: error as fraction of the multiple,
	// consistent with removeMultiples approach
	if absFloat(ratio-float64(rounded))/float64(rounded) > episodeRatioTolerance {
		return 1
	}

	return rounded
}

func FormatDuration(secs int) string {
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

// ptsDuration calculates duration in seconds from two PTS timestamps.
// uint32 subtraction handles PTS wraparound at 0xFFFFFFFF naturally.
func ptsDuration(in, out uint32) int {
	return int((out - in) / ptsClock)
}

func ParsePlaylist(path string) (*Playlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	typeInd := make([]byte, 4)
	io.ReadFull(f, typeInd)
	if string(typeInd) != typeIndicatorMPLS {
		return nil, fmt.Errorf("not an mpls file")
	}

	// Read PlayList and PlayListMark offsets from header
	f.Seek(playlistOffsetAddr, io.SeekStart)
	var playlistOffset uint32
	binary.Read(f, binary.BigEndian, &playlistOffset)

	f.Seek(playlistMarkOffsetAddr, io.SeekStart)
	var markOffset uint32
	binary.Read(f, binary.BigEndian, &markOffset)

	// PlayList() header: length(4) + reserved(2) + num_items(2) + num_subpaths(2)
	f.Seek(int64(playlistOffset)+playlistHeaderSkip, io.SeekStart)
	var numItems uint16
	binary.Read(f, binary.BigEndian, &numItems)
	f.Seek(2, io.SeekCurrent) // skip num_subpaths

	pl := &Playlist{
		Name:      strings.TrimSuffix(filepath.Base(path), ".mpls"),
		PlayItems: make([]PlayItem, 0, numItems),
	}

	for i := 0; i < int(numItems); i++ {
		itemStart, _ := f.Seek(0, io.SeekCurrent)

		var itemLen uint16
		binary.Read(f, binary.BigEndian, &itemLen)

		clipName := make([]byte, playItemClipNameLen)
		io.ReadFull(f, clipName)

		f.Seek(playItemTimestampSkip, io.SeekCurrent)

		var inTime, outTime uint32
		binary.Read(f, binary.BigEndian, &inTime)
		binary.Read(f, binary.BigEndian, &outTime)

		pl.PlayItems = append(pl.PlayItems, PlayItem{
			ClipName: string(clipName[:playItemClipNameUsed]),
			InTime:   inTime,
			OutTime:  outTime,
			Duration: ptsDuration(inTime, outTime),
		})

		// Jump to next PlayItem using itemLen — handles variable-length
		// stream entries regardless of audio/subtitle track count.
		f.Seek(itemStart+2+int64(itemLen), io.SeekStart)
	}

	// Parse PlayListMark section
	if markOffset > 0 {
		pl.Marks, _ = parsePlaylistMarks(f, markOffset)
	}

	return pl, nil
}

func parsePlaylistMarks(f *os.File, markOffset uint32) ([]PlaylistMark, error) {
	f.Seek(int64(markOffset), io.SeekStart)

	var length uint32
	binary.Read(f, binary.BigEndian, &length)

	var numMarks uint16
	binary.Read(f, binary.BigEndian, &numMarks)

	marks := make([]PlaylistMark, numMarks)
	for i := range marks {
		var markType uint8
		var playItemRef uint16
		var reserved uint16
		var ts uint32
		var esPid uint16
		var dur uint32

		binary.Read(f, binary.BigEndian, &markType)
		binary.Read(f, binary.BigEndian, &playItemRef)
		binary.Read(f, binary.BigEndian, &reserved)
		binary.Read(f, binary.BigEndian, &ts)
		binary.Read(f, binary.BigEndian, &esPid)
		binary.Read(f, binary.BigEndian, &dur)

		marks[i] = PlaylistMark{
			MarkType:    markType,
			PlayItemRef: playItemRef,
			Timestamp:   ts,
			Duration:    dur,
		}
	}
	return marks, nil
}

func LoadAllPlaylists(bdmvRoot string) ([]*Playlist, error) {
	pattern := filepath.Join(bdmvRoot, "PLAYLIST", "*.mpls")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var playlists []*Playlist
	for _, f := range files {
		pl, err := ParsePlaylist(f)
		if err != nil {
			continue
		}
		playlists = append(playlists, pl)
	}

	sort.Slice(playlists, func(i, j int) bool {
		return playlists[i].Name < playlists[j].Name
	})

	return playlists, nil
}

func LoadEpisodePlaylists(bdmvRoot string, minDur, maxDur, clusterDur int) ([]*Playlist, error) {
	all, err := LoadAllPlaylists(bdmvRoot)
	if err != nil {
		return nil, err
	}

	var episodes []*Playlist
	seen := map[string]bool{}

	for _, pl := range all {
		clip := pl.PrimaryClip()
		if clip == "" || seen[clip] {
			continue
		}

		dur := pl.EstimateDuration(bdmvRoot)
		if dur < minViableDuration {
			continue
		}

		episodeCount := pl.EstimateEpisodeCount(bdmvRoot, clusterDur)
		perEpisodeDur := dur / episodeCount

		if perEpisodeDur < minDur || perEpisodeDur > maxDur {
			continue
		}

		seen[clip] = true

		if episodeCount > 1 {
			for i := 0; i < episodeCount; i++ {
				episodes = append(episodes, &Playlist{
					Name:      fmt.Sprintf("%s[%d]", pl.Name, i+1),
					PlayItems: pl.PlayItems,
					Marks:     pl.Marks,
				})
			}
		} else {
			episodes = append(episodes, pl)
		}
	}

	return episodes, nil
}
