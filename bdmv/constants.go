package bdmv

const (
	// BDMV file type indicators
	TypeIndicatorINDX = "INDX"
	TypeIndicatorMOBJ = "MOBJ"
	TypeIndicatorMPLS = "MPLS"

	// index.bdmv parsing
	indexAppInfoBodyOffset = 4 + 16 // length(4) + reserved(16)

	// Title entry object types
	ObjectTypeHDMVFirstPlay = 0x40
	ObjectTypeHDMV          = 0xA0
	ObjectTypeBDJ           = 0x60
	ObjectTypeMask          = 0xE0

	// MovieObject.bdmv parsing
	movieObjectTableOffset = 0x28

	// MovieObject flags
	mobjFlagResumeIntention = 0x8000
	mobjFlagMenuCallMask    = 0x4000
	mobjFlagTitleSearchMask = 0x2000

	// Navigation command size in bytes
	navCommandSize = 12

	// Playlist parsing
	playlistOffsetAddr     = 0x08
	playlistMarkOffsetAddr = 0x0C
	playlistHeaderSkip     = 4 + 2 // length(4) + reserved(2)
	playItemClipNameLen    = 9
	playItemClipNameUsed   = 5 // just the number part e.g. "01061"
	playItemTimestampSkip  = 4 // reserved(1) + connection(1) + stc_id(1) + reserved(1)

	// PTS clock rate for Blu-ray (45kHz)
	PTSClock = 45000
)
