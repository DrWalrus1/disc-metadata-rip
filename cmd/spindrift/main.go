package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/DrWalrus1/spindrift/bdmv"
	"github.com/DrWalrus1/spindrift/disc"
	"github.com/DrWalrus1/spindrift/env"
	"github.com/DrWalrus1/spindrift/tmdb"
)

const envTMDBAPIKey = "TMDB_API_KEY"

func run(bdmvRoot string, startEpisode int) error {
	d, err := disc.Open(bdmvRoot)
	if err != nil {
		return err
	}

	// --- index.bdmv ---
	fmt.Printf("Version:    %s\n", d.Index.Version)
	fmt.Printf("FirstPlay → %s\n", d.Index.FirstPlay.PlaylistPath(bdmvRoot))
	fmt.Printf("TopMenu   → %s\n", d.Index.TopMenu.PlaylistPath(bdmvRoot))
	for i, t := range d.Index.Titles {
		if t.IsHDMV() {
			fmt.Printf("  Title[%d] → PLAYLIST/%05d.mpls\n", i, t.ObjectIDRef)
		} else {
			fmt.Printf("  Title[%d] → MovieObject[%d] (BD-J)\n", i, t.ObjectIDRef)
		}
	}

	// --- MovieObject.bdmv ---
	fmt.Printf("\nObjects (%d):\n", len(d.MObj.Objects))
	for i, obj := range d.MObj.Objects {
		fmt.Printf("  [%d] resume=%v menuMask=%v titleMask=%v cmds=%d\n",
			i, obj.ResumeIntentionFlag, obj.MenuCallMask, obj.TitleSearchMask,
			len(obj.Commands),
		)
	}

	// --- Disc metadata ---
	fmt.Printf("\nDisc Title: %s\n", d.Info.ShowName)
	if !d.Info.IsMovie {
		fmt.Printf("Season:     %d\n", d.Info.Season)
		fmt.Printf("Disc:       %d\n", d.Info.Disc)
	}
	if startEpisode > 0 {
		fmt.Printf("Start Ep:   %d\n", startEpisode)
	}
	fmt.Println()

	// --- Infer episode duration bounds ---
	minDur, maxDur, clusterDur := disc.InferEpisodeBounds(bdmvRoot)
	fmt.Printf("Inferred episode bounds: %s – %s (cluster center: %s)\n",
		bdmv.FormatDuration(minDur),
		bdmv.FormatDuration(maxDur),
		bdmv.FormatDuration(clusterDur),
	)

	// --- Episode playlists ---
	episodes, err := disc.LoadEpisodePlaylists(bdmvRoot, minDur, maxDur, clusterDur)
	if err != nil {
		return fmt.Errorf("loading playlists: %w", err)
	}

	if len(episodes) == 0 {
		fmt.Println("No episodes found on disc")
		return nil
	}

	d.Info.DetectMovie(len(episodes))

	if d.Info.IsMovie {
		fmt.Printf("Detected: Movie\n\n")
	} else {
		fmt.Printf("Found %d episodes on disc\n\n", len(episodes))
	}

	// --- TMDB lookup ---
	apiKey := os.Getenv(envTMDBAPIKey)
	if apiKey == "" {
		return fmt.Errorf("%s not set — add it to .env or your environment", envTMDBAPIKey)
	}

	client := tmdb.New(apiKey)

	if d.Info.IsMovie {
		return runMovie(client, episodes, d.Info, bdmvRoot, clusterDur)
	}
	return runTV(client, episodes, d.Info, bdmvRoot, clusterDur, startEpisode)
}

func runMovie(
	client *tmdb.Client,
	episodes []*bdmv.Playlist,
	info disc.DiscInfo,
	bdmvRoot string,
	clusterDur int,
) error {
	movies, err := client.SearchMovie(info.ShowName)
	if err != nil {
		return fmt.Errorf("TMDB movie search: %w", err)
	}
	if len(movies) == 0 {
		fmt.Printf("No TMDB movie results found for %q\n", info.ShowName)
		printMovieNoTMDB(episodes, bdmvRoot, clusterDur)
		return nil
	}

	movie := movies[0]
	details, err := client.GetMovie(movie.ID)
	if err != nil {
		return fmt.Errorf("TMDB movie details: %w", err)
	}

	fmt.Printf("TMDB Match: %s (ID: %d, Runtime: %d min)\n\n",
		details.Title, movie.ID, details.Runtime)

	printMovie(episodes[0], details, bdmvRoot, clusterDur)
	return nil
}

func runTV(
	client *tmdb.Client,
	episodes []*bdmv.Playlist,
	info disc.DiscInfo,
	bdmvRoot string,
	clusterDur int,
	startEpisode int,
) error {
	shows, err := client.SearchTV(info.ShowName)
	if err != nil {
		return fmt.Errorf("TMDB search: %w", err)
	}
	if len(shows) == 0 {
		fmt.Printf("No TMDB results found for %q\n", info.ShowName)
		printEpisodesNoTMDB(episodes, info, bdmvRoot, clusterDur, startEpisode)
		return nil
	}

	show := shows[0]
	fmt.Printf("TMDB Match: %s (ID: %d)\n\n", show.Name, show.ID)

	season, err := client.GetSeason(show.ID, info.Season)
	if err != nil {
		return fmt.Errorf("TMDB season fetch: %w", err)
	}

	tmdbEps := tmdb.EpisodesForDisc(season, startEpisode, len(episodes))
	printEpisodes(episodes, tmdbEps, info, bdmvRoot, clusterDur)
	return nil
}

func printMovie(
	pl *bdmv.Playlist,
	details *tmdb.MovieDetails,
	bdmvRoot string,
	clusterDur int,
) {
	dur := pl.EstimateDuration(bdmvRoot, disc.DefaultBitrate)
	count := disc.EstimateEpisodeCount(pl, bdmvRoot, clusterDur)

	fmt.Printf("%-10s %-12s %-12s %s\n", "Type", "Playlist", "Duration", "Title")
	fmt.Println(strings.Repeat("-", 55))
	fmt.Printf("%-10s %-12s %-12s %s\n",
		"Movie",
		pl.Name,
		bdmv.FormatDuration(dur/count),
		details.Title,
	)
}

func printMovieNoTMDB(
	episodes []*bdmv.Playlist,
	bdmvRoot string,
	clusterDur int,
) {
	fmt.Printf("%-10s %-12s %s\n", "Type", "Playlist", "Duration")
	fmt.Println(strings.Repeat("-", 35))
	for _, pl := range episodes {
		dur := pl.EstimateDuration(bdmvRoot, disc.DefaultBitrate)
		count := disc.EstimateEpisodeCount(pl, bdmvRoot, clusterDur)
		fmt.Printf("%-10s %-12s %s\n",
			"Movie",
			pl.Name,
			bdmv.FormatDuration(dur/count),
		)
	}
}

func printEpisodes(
	episodes []*bdmv.Playlist,
	tmdbEps []tmdb.Episode,
	info disc.DiscInfo,
	bdmvRoot string,
	clusterDur int,
) {
	fmt.Printf("%-6s %-14s %-10s %-12s %s\n",
		"Ep", "Playlist", "Clip", "Duration", "Title")
	fmt.Println(strings.Repeat("-", 72))

	for i, pl := range episodes {
		dur := pl.EstimateDuration(bdmvRoot, disc.DefaultBitrate)
		count := disc.EstimateEpisodeCount(pl, bdmvRoot, clusterDur)

		epLabel := fmt.Sprintf("S%02dE??", info.Season)
		title := "unknown"
		if i < len(tmdbEps) {
			ep := tmdbEps[i]
			epLabel = fmt.Sprintf("S%02dE%02d", info.Season, ep.EpisodeNumber)
			title = ep.Name
		}

		fmt.Printf("%-6s %-14s %-10s %-12s %s\n",
			epLabel,
			pl.Name,
			pl.PrimaryClip(),
			bdmv.FormatDuration(dur/count),
			title,
		)
	}
}

func printEpisodesNoTMDB(
	episodes []*bdmv.Playlist,
	info disc.DiscInfo,
	bdmvRoot string,
	clusterDur int,
	startEpisode int,
) {
	fmt.Printf("%-6s %-14s %-10s %s\n", "Ep", "Playlist", "Clip", "Duration")
	fmt.Println(strings.Repeat("-", 48))

	for i, pl := range episodes {
		dur := pl.EstimateDuration(bdmvRoot, disc.DefaultBitrate)
		count := disc.EstimateEpisodeCount(pl, bdmvRoot, clusterDur)

		epNum := startEpisode + i
		if startEpisode == 0 {
			epNum = i + 1
		}

		fmt.Printf("S%02dE%02d %-14s %-10s %s\n",
			info.Season,
			epNum,
			pl.Name,
			pl.PrimaryClip(),
			bdmv.FormatDuration(dur/count),
		)
	}
}

func parseArgs() (discArg string, startEpisode int) {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--start-episode":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &startEpisode)
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && discArg == "" {
				discArg = args[i]
			}
		}
	}
	return discArg, startEpisode
}

func main() {
	if err := env.Load(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	discArg, startEpisode := parseArgs()

	bdmvRoot, err := disc.SelectBDMV(discArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Usage: spindrift [disc-path] [--start-episode N]\n")
		os.Exit(1)
	}

	fmt.Printf("Found disc: %s\n", bdmvRoot)

	if err := run(bdmvRoot, startEpisode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
