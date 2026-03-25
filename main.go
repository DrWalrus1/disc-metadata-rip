package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const envTMDBAPIKey = "TMDB_API_KEY"

func parseDiscTitle(bdmvRoot string) (string, error) {
	path := filepath.Join(bdmvRoot, bdmtEnglishXML)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	start := strings.Index(string(data), diNameOpenTag)
	end := strings.Index(string(data), diNameCloseTag)
	if start < 0 || end < 0 {
		return "", fmt.Errorf("could not find disc name in XML")
	}

	name := string(data)[start+len(diNameOpenTag) : end]
	return strings.Join(strings.Fields(name), " "), nil
}

func run(bdmvRoot string) error {
	// --- index.bdmv ---
	idx, err := ParseIndex(filepath.Join(bdmvRoot, "index.bdmv"))
	if err != nil {
		return fmt.Errorf("parsing index.bdmv: %w", err)
	}

	fmt.Printf("Version:    %s\n", idx.Version)
	fmt.Printf("FirstPlay → %s\n", idx.FirstPlay.PlaylistPath(bdmvRoot))
	fmt.Printf("TopMenu   → %s\n", idx.TopMenu.PlaylistPath(bdmvRoot))
	for i, t := range idx.Titles {
		if t.IsHDMV() {
			fmt.Printf("  Title[%d] → PLAYLIST/%05d.mpls\n", i, t.ObjectIDRef)
		} else {
			fmt.Printf("  Title[%d] → MovieObject[%d] (BD-J)\n", i, t.ObjectIDRef)
		}
	}

	// --- MovieObject.bdmv ---
	mobj, err := ParseMovieObject(filepath.Join(bdmvRoot, "MovieObject.bdmv"))
	if err != nil {
		return fmt.Errorf("parsing MovieObject.bdmv: %w", err)
	}

	fmt.Printf("\nVersion: %s\n", mobj.Version)
	fmt.Printf("Objects (%d):\n", len(mobj.Objects))
	for i, obj := range mobj.Objects {
		fmt.Printf("  [%d] resume=%v menuMask=%v titleMask=%v cmds=%d\n",
			i, obj.ResumeIntentionFlag, obj.MenuCallMask, obj.TitleSearchMask,
			len(obj.Commands),
		)
	}

	// --- Disc metadata ---
	discTitle, err := parseDiscTitle(bdmvRoot)
	if err != nil {
		fmt.Printf("Warning: could not parse disc title: %v\n", err)
		discTitle = filepath.Base(filepath.Dir(bdmvRoot))
		fmt.Printf("Falling back to volume name: %s\n", discTitle)
	}
	fmt.Printf("\nDisc Title: %s\n", discTitle)

	discInfo := ParseDiscTitle(discTitle)
	fmt.Printf("Show:       %s\n", discInfo.ShowName)
	fmt.Printf("Season:     %d\n", discInfo.Season)
	fmt.Printf("Disc:       %d\n\n", discInfo.Disc)

	// --- Infer episode duration bounds from disc ---
	minDur, maxDur := InferEpisodeBounds(bdmvRoot)

	// --- Episode playlists ---
	episodes, err := LoadEpisodePlaylists(bdmvRoot, minDur, maxDur)
	if err != nil {
		return fmt.Errorf("loading playlists: %w", err)
	}

	if len(episodes) == 0 {
		fmt.Println("No episodes found on disc")
		return nil
	}
	fmt.Printf("Found %d episodes on disc\n\n", len(episodes))

	// --- TMDB lookup ---
	apiKey := os.Getenv(envTMDBAPIKey)
	if apiKey == "" {
		return fmt.Errorf("%s not set — add it to .env or your environment", envTMDBAPIKey)
	}

	client := NewTMDBClient(apiKey)

	shows, err := client.SearchTV(discInfo.ShowName)
	if err != nil {
		return fmt.Errorf("TMDB search: %w", err)
	}
	if len(shows) == 0 {
		fmt.Printf("No TMDB results found for %q\n", discInfo.ShowName)
		printEpisodesNoTMDB(episodes, discInfo, bdmvRoot)
		return nil
	}

	show := shows[0]
	fmt.Printf("TMDB Match: %s (ID: %d)\n\n", show.Name, show.ID)

	season, err := client.GetSeason(show.ID, discInfo.Season)
	if err != nil {
		return fmt.Errorf("TMDB season fetch: %w", err)
	}

	tmdbEps := EpisodesForDisc(season, discInfo.Disc, len(episodes))

	// --- Results ---
	printEpisodes(episodes, tmdbEps, discInfo, bdmvRoot)
	return nil
}

func printEpisodes(
	episodes []*Playlist,
	tmdbEps []TMDBEpisode,
	discInfo DiscInfo,
	bdmvRoot string,
) {
	fmt.Printf("%-6s %-12s %-10s %-12s %s\n",
		"Ep", "Playlist", "Clip", "Duration", "Title")
	fmt.Println(strings.Repeat("-", 70))

	for i, pl := range episodes {
		epNum := (discInfo.Disc-1)*len(episodes) + i + 1
		title := "unknown"
		if i < len(tmdbEps) {
			title = tmdbEps[i].Name
		}
		fmt.Printf("S%02dE%02d %-12s %-10s %-12s %s\n",
			discInfo.Season,
			epNum,
			pl.Name,
			pl.PrimaryClip(),
			FormatDuration(pl.EstimateDuration(bdmvRoot)),
			title,
		)
	}
}

func printEpisodesNoTMDB(
	episodes []*Playlist,
	discInfo DiscInfo,
	bdmvRoot string,
) {
	fmt.Printf("%-6s %-12s %-10s %s\n", "Ep", "Playlist", "Clip", "Duration")
	fmt.Println(strings.Repeat("-", 45))

	for i, pl := range episodes {
		epNum := (discInfo.Disc-1)*len(episodes) + i + 1
		fmt.Printf("S%02dE%02d %-12s %-10s %s\n",
			discInfo.Season,
			epNum,
			pl.Name,
			pl.PrimaryClip(),
			FormatDuration(pl.EstimateDuration(bdmvRoot)),
		)
	}
}

func main() {
	// Load .env from current working directory
	if err := loadEnv(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	var discArg string
	if len(os.Args) > 1 {
		discArg = os.Args[1]
	}

	bdmvRoot, err := SelectBDMV(discArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Usage: %s [disc-path]\n", os.Args[0])
		os.Exit(1)
	}

	if err := run(bdmvRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
