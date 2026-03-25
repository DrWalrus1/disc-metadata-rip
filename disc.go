package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// discSearchRoots returns the platform-specific directories to search
// for mounted optical disc volumes.
func discSearchRoots() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/Volumes"}
	case "linux":
		roots := []string{"/media", "/run/media", "/mnt"}
		// /run/media/<username>/<disc> — expand one level for user dirs
		var expanded []string
		for _, root := range roots {
			entries, err := os.ReadDir(root)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					expanded = append(expanded, filepath.Join(root, e.Name()))
				}
			}
		}
		return expanded
	default:
		return nil
	}
}

// FindBDMVRoots searches mounted volumes for BDMV directories and
// returns all found paths.
func FindBDMVRoots() ([]string, error) {
	roots := discSearchRoots()
	if len(roots) == 0 {
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	var found []string
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			bdmv := filepath.Join(root, e.Name(), "BDMV")
			if _, err := os.Stat(bdmv); err == nil {
				found = append(found, bdmv)
			}
		}
	}

	return found, nil
}

// SelectBDMV returns a single BDMV root, either from the command line
// argument, auto-detected, or interactively chosen if multiple are found.
func SelectBDMV(arg string) (string, error) {
	// Explicit path provided
	if arg != "" {
		// Accept either the BDMV dir itself or its parent
		if filepath.Base(arg) == "BDMV" {
			if _, err := os.Stat(arg); err == nil {
				return arg, nil
			}
		}
		bdmv := filepath.Join(arg, "BDMV")
		if _, err := os.Stat(bdmv); err == nil {
			return bdmv, nil
		}
		return "", fmt.Errorf("no BDMV directory found at %s", arg)
	}

	// Auto-detect
	found, err := FindBDMVRoots()
	if err != nil {
		return "", err
	}

	switch len(found) {
	case 0:
		return "", fmt.Errorf("no Blu-ray disc found — is a disc mounted?")
	case 1:
		fmt.Printf("Found disc: %s\n", filepath.Dir(found[0]))
		return found[0], nil
	default:
		fmt.Println("Multiple discs found:")
		for i, f := range found {
			fmt.Printf("  [%d] %s\n", i+1, filepath.Dir(f))
		}
		fmt.Print("Select disc [1]: ")
		choice := 1
		fmt.Scan(&choice)
		if choice < 1 || choice > len(found) {
			return "", fmt.Errorf("invalid selection")
		}
		return found[choice-1], nil
	}
}

// streamDurations returns estimated durations for all unique clips
// above a minimum viable size (to exclude obvious stubs).
func streamDurations(bdmvRoot string) []int {
	pattern := filepath.Join(bdmvRoot, "PLAYLIST", "*.mpls")
	files, _ := filepath.Glob(pattern)

	seen := map[string]bool{}
	var durations []int

	for _, f := range files {
		pl, err := ParsePlaylist(f)
		if err != nil {
			continue
		}
		clip := pl.PrimaryClip()
		if clip == "" || seen[clip] {
			continue
		}
		seen[clip] = true

		dur := pl.EstimateDuration(bdmvRoot)
		if dur >= minViableDuration {
			durations = append(durations, dur)
		}
	}

	sort.Ints(durations)
	return durations
}

// dominantCluster finds the largest cluster of similar durations
// using a sliding window approach. Returns (min, max) bounds.
func dominantCluster(durations []int) (min, max int) {
	if len(durations) == 0 {
		return minEpisodeDuration, maxEpisodeDuration
	}

	if len(durations) == 1 {
		d := durations[0]
		return int(float64(d) * clusterLowerBound),
			int(float64(d) * clusterUpperBound)
	}

	// Find the window with the most values where all values are within
	// clusterTolerance of each other
	bestStart, bestCount := 0, 1

	for i := 0; i < len(durations); i++ {
		count := 1
		for j := i + 1; j < len(durations); j++ {
			ratio := float64(durations[j]) / float64(durations[i])
			if ratio <= clusterTolerance {
				count++
			} else {
				break
			}
		}
		if count > bestCount {
			bestCount = count
			bestStart = i
		}
	}

	clusterMin := durations[bestStart]
	clusterMax := durations[bestStart+bestCount-1]

	// Expand bounds by tolerance margin to catch slight variations
	return int(float64(clusterMin) * clusterLowerBound),
		int(float64(clusterMax) * clusterUpperBound)
}

// InferEpisodeBounds analyses stream durations on the disc and returns
// likely min/max episode duration bounds in seconds.
func InferEpisodeBounds(bdmvRoot string) (min, max int) {
	durations := streamDurations(bdmvRoot)

	if len(durations) == 0 {
		return minEpisodeDuration, maxEpisodeDuration
	}

	min, max = dominantCluster(durations)

	fmt.Printf("Inferred episode bounds: %s – %s",
		FormatDuration(min), FormatDuration(max))
	fmt.Printf(" (from %d unique stream durations)\n", len(durations))

	return min, max
}
