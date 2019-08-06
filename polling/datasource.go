package polling

// IDataSource is the implementation all data sources need to adheer to.
type IDataSource interface {
	// Include some human friendly things.
	Name() string // Human friendly name
	Url() string  // Url to their website

	// FetchPegPrices is a rest based API call to the data source to fetch
	// the prices for the supported pegs.
	FetchPegPrices() (peg PegAssets, err error)

	// FetchPegPrice only fetches the price for a single peg
	FetchPegPrice(peg string) (PegItem, error)

	// SupportedPegs tells us what supported pegs the exchange supports.
	// All exchanges should have a list of pegs they support. This should
	// be defined up front.
	SupportedPegs() []string
}
