package asset_fee_liquidity

type AssetFeeLiquiditiesVersion uint64

const (
	SIMPLE AssetFeeLiquiditiesVersion = iota
)

func (t AssetFeeLiquiditiesVersion) String() string {
	switch t {
	case SIMPLE:
		return "SIMPLE"
	default:
		return "Unknown AssetFeeLiquiditiesVersion"
	}
}
