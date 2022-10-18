package util

import "math/big"

// GetCoinOutWithFees Swapping x -> y
// dx_f = dx(1-fee)
// (x + dx_f)*(y - dy) = x*y
// dy = y * dx_f / (x + dx_f)
func GetCoinOutWithFees(coinInVal *big.Int, reserveInSize *big.Int, reserveOutSize *big.Int, feeBps, feeScale int64) *big.Int {
	feePct := big.NewInt(feeBps)
	scale := big.NewInt(feeScale)
	feeMultiplier := big.NewInt(0).Sub(scale, feePct)
	coinInAfterFees := big.NewInt(0).Mul(coinInVal, feeMultiplier)
	newReservesInSize := big.NewInt(0).Add(
		big.NewInt(0).Mul(
			reserveInSize,
			scale,
		),
		coinInAfterFees,
	)

	return big.NewInt(0).Div(
		big.NewInt(0).Mul(
			coinInAfterFees,
			reserveOutSize,
		),
		newReservesInSize,
	)
}
