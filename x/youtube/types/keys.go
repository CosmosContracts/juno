package types

var (
	ParamsKey  = []byte{0x00}
	VideoKey   = []byte{0x01} // holds the Title, Summary, and list of video IDs
	ContentKey = []byte{0x02} // holds the actual video content indexed by video IDs
)

const (
	ModuleName = "youtube"

	StoreKey = ModuleName

	QuerierRoute = ModuleName
)
