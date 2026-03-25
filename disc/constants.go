package disc

const (
	// Bitrate estimate for duration calculation (35 Mbps)
	DefaultBitrate = 35_000_000

	// Minimum stream duration to consider for clustering
	MinViableDuration = 2 * 60 // 2 minutes

	// Minimum duration to include in cluster analysis
	MinClusterDuration = 10 * 60 // 10 minutes

	// Maximum plausible episodes in one stream
	MaxEpisodesPerStream = 8

	// Relative tolerance for episode count detection
	EpisodeRatioTolerance = 0.15

	// Episode duration filter fallback
	MinEpisodeDuration = 18 * 60 // 18 minutes
	MaxEpisodeDuration = 50 * 60 // 50 minutes

	// Cluster detection tolerance
	ClusterTolerance = 1.30

	// Cluster bound expansion margins
	ClusterLowerBound = 0.80
	ClusterUpperBound = 1.20

	// Multiple detection settings
	MultipleDetectionTolerance = 0.15
	MinMultiple                = 3.0

	// Disc metadata XML tags
	DiNameOpenTag  = "<di:name>"
	DiNameCloseTag = "</di:name>"

	// Disc metadata path relative to BDMV root
	BDMTEnglishXML = "META/DL/bdmt_eng.xml"
)
