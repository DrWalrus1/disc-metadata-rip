package main

const (
	// BDMV file type indicators
	typeIndicatorINDX = "INDX"
	typeIndicatorMOBJ = "MOBJ"
	typeIndicatorMPLS = "MPLS"

	// index.bdmv parsing
	indexAppInfoBodyOffset = 4 + 16 // length(4) + reserved(16)

	// Title entry object types
	objectTypeHDMVFirstPlay = 0x40
	objectTypeHDMV          = 0xA0
	objectTypeBDJ           = 0x60
	objectTypeMask          = 0xE0

	// MovieObject.bdmv parsing
	movieObjectTableOffset = 0x28

	// MovieObject flags
	mobjFlagResumeIntention = 0x8000
	mobjFlagMenuCallMask    = 0x4000
	mobjFlagTitleSearchMask = 0x2000

	// Navigation command size in bytes
	navCommandSize = 12

	// playlist.bdmv parsing
	playlistOffsetAddr     = 0x08
	playlistMarkOffsetAddr = 0x0C
	playlistHeaderSkip     = 4 + 2 // length(4) + reserved(2)
	playItemClipNameLen    = 9
	playItemClipNameUsed   = 5 // just the number part e.g. "01061"
	playItemTimestampSkip  = 4 // reserved(1) + connection(1) + stc_id(1) + reserved(1)

	// PTS clock rate for Blu-ray (45kHz)
	ptsClock = 45000

	// Bitrate estimate for duration calculation (35 Mbps)
	estimatedBitrate = 35_000_000

	// Minimum stream duration to consider for clustering
	minViableDuration = 2 * 60 // 2 minutes

	// Minimum duration to include in cluster analysis —
	// filters out bumpers and short clips that skew the cluster
	minClusterDuration = 10 * 60 // 10 minutes

	// Episode count detection — maximum plausible episodes in one stream
	maxEpisodesPerStream = 8

	// Episode count detection — relative tolerance per multiple
	// e.g. 0.15 = within 15% of the nearest integer multiple
	episodeRatioTolerance = 0.15

	// Episode duration filter fallback (used when inference fails)
	minEpisodeDuration = 18 * 60 // 18 minutes
	maxEpisodeDuration = 50 * 60 // 50 minutes

	// Cluster detection tolerance — streams within this ratio are
	// considered the same "type" (e.g. 1.3 = within 30% of each other)
	clusterTolerance = 1.30

	// Expand the inferred cluster bounds by this margin
	clusterLowerBound = 0.80 // 20% below cluster min
	clusterUpperBound = 1.20 // 20% above cluster max

	// How close a duration must be to a multiple to be considered one
	// expressed as a fraction of the multiple (relative tolerance)
	multipleDetectionTolerance = 0.15

	// Minimum multiple to consider as a "play all" stream (3x or more)
	minMultiple = 3.0

	// TMDB
	tmdbBaseURL = "https://api.themoviedb.org/3"

	// Disc metadata XML tags
	diNameOpenTag  = "<di:name>"
	diNameCloseTag = "</di:name>"

	// Disc metadata path
	bdmtEnglishXML = "META/DL/bdmt_eng.xml"

	// Environment variable names
	envTMDBAPIKey = "TMDB_API_KEY"
)
