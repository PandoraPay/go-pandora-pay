package asset_fee_liquidity

type AssetFeeLiquiditiesVersion uint64

const (
	NONE AssetFeeLiquiditiesVersion = iota
	SIMPLE
)

func (t AssetFeeLiquiditiesVersion) String() string {
	switch t {
	case SIMPLE:
		return "SIMPLE"
	default:
		return "Unknown AssetFeeLiquiditiesVersion"
	}
}

type UpdateLiquidityStatus byte

const (
	UPDATE_LIQUIDITY_NOTHING = iota
	UPDATE_LIQUIDITY_OVERWRITTEN
	UPDATE_LIQUIDITY_INSERTED
	UPDATE_LIQUIDITY_DELETED
)
