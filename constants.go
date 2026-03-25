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
	playlistOffsetAddr    = 0x08
	playlistHeaderSkip    = 4 + 2 // length(4) + reserved(2)
	playItemClipNameLen   = 9
	playItemClipNameUsed  = 5 // just the number part e.g. "01061"
	playItemTimestampSkip = 4 // reserved(1) + connection(1) + stc_id(1) + reserved(1)

	// PTS clock rate for Blu-ray (45kHz)
	ptsClock = 45000

	// Stream file size thresholds
	minEpisodeBytes = 500 * 1024 * 1024 // 500 MB

	// Bitrate estimate for duration calculation (35 Mbps)
	estimatedBitrate = 35_000_000

	// Episode duration filter (in seconds)
	minEpisodeDuration = 18 * 60 // 18 minutes
	maxEpisodeDuration = 50 * 60 // 50 minutes

	// TMDB
	tmdbBaseURL = "https://api.themoviedb.org/3"

	// Disc metadata XML tags
	diNameOpenTag  = "<di:name>"
	diNameCloseTag = "</di:name>"

	// Disc metadata path
	bdmtEnglishXML = "META/DL/bdmt_eng.xml"

	// Minimum stream size to consider for duration clustering (filters stubs)
	minViableDuration = 2 * 60 // 2 minutes

	// Cluster detection tolerance — streams within this ratio are
	// considered the same "type" (e.g. 1.3 = within 30% of each other)
	clusterTolerance = 1.30

	// Expand the inferred cluster bounds by this margin
	clusterLowerBound = 0.80 // 20% below cluster min
	clusterUpperBound = 1.20 // 20% above cluster max
)
