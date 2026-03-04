package utils

import (
	"math/big"
)

// MaxSwapFee defines the maximum swap fee in hundredths of a bip (1e6 = 100%).
var (
	MaxSwapFee = new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
)

// ComputeSwapStep calculates the result of a single swap step within a tick.
//
// This function is a Go translation of the Uniswap V3 swap step calculation,
// handling both exact input and exact output swaps, taking liquidity, sqrt prices,
// and swap fees into account.
//
// Parameters:
//   - sqrtPriceCurrentX96: current sqrt price in Q64.96 format
//   - sqrtPriceTargetX96: target sqrt price for this swap step (usually the next initialized tick)
//   - liquidity: current in-range liquidity
//   - amountRemaining: amount left to swap (positive for exact input, negative for exact output)
//   - feePips: swap fee in hundredths of a bip (1e6 = 100%)
//
// Returns:
//   - sqrtPriceNextX96: the sqrt price after this swap step
//   - amountIn: amount of input token consumed
//   - amountOut: amount of output token produced
//   - feeAmount: swap fee taken in input token units
//   - err: error if any computation fails
//
// Behavior:
//   - Determines the swap direction (zeroForOne) based on current and target sqrt prices.
//   - For exact input swaps:
//   - Deducts the fee from the remaining input amount
//   - Checks if target price can be reached with remaining input
//   - If not, computes next sqrt price using GetNextSqrtPriceFromInput
//   - For exact output swaps:
//   - Checks if target price can be reached with remaining output
//   - If not, computes next sqrt price using GetNextSqrtPriceFromOutput
//   - Computes the input/output amounts based on liquidity and swap direction
//   - Calculates the fee amount depending on whether target price was fully reached
//
// Notes:
//   - Uses big.Int arithmetic to handle Q64.96 fixed-point numbers and large liquidity values
//   - Implements rounding up for fee calculations via MulDivRoundingUp
//   - Mirrors Uniswap V3 core swap step logic
func ComputeSwapStep(sqrtPriceCurrentX96, sqrtPriceTargetX96, liquidity, amountRemaining *big.Int, feePips uint32) (sqrtPriceNextX96, amountIn, amountOut, feeAmount *big.Int, err error) {
	sqrtPriceNextX96 = new(big.Int)
	amountIn = new(big.Int)
	amountOut = new(big.Int)
	feeAmount = new(big.Int)

	_feePips := feePips
	zeroForOne := sqrtPriceCurrentX96.Cmp(sqrtPriceTargetX96) >= 0
	exactIn := amountRemaining.Sign() < 0

	if exactIn {
		// deduct fee from remaining input
		amountRemainingLessFee := new(big.Int)
		amountRemainingLessFee, err = MulDiv(new(big.Int).Abs(amountRemaining), new(big.Int).Sub(MaxSwapFee, big.NewInt(int64(feePips))), MaxSwapFee)
		if err != nil {
			return
		}
		if zeroForOne {
			amountIn, err = GetAmount0Delta(sqrtPriceTargetX96, sqrtPriceCurrentX96, liquidity, true)
			if err != nil {
				return
			}
		} else {
			amountIn, err = GetAmount1Delta(sqrtPriceCurrentX96, sqrtPriceTargetX96, liquidity, true)
			if err != nil {
				return
			}
		}
		if amountRemainingLessFee.Cmp(amountIn) >= 0 {
			sqrtPriceNextX96 = sqrtPriceTargetX96
			if int64(_feePips) == MaxSwapFee.Int64() {
				feeAmount.Set(amountIn)
			} else {
				tmp := new(big.Int)
				denominator := new(big.Int).Sub(MaxSwapFee, big.NewInt(int64(_feePips)))
				tmp, err = MulDivRoundingUp(amountIn, big.NewInt(int64(_feePips)), denominator)
				if err != nil {
					return
				}
				feeAmount.Set(tmp)
			}
		} else {
			// exhaust the remaining amount
			amountIn.Set(amountRemainingLessFee)
			sqrtPriceNextX96, err = GetNextSqrtPriceFromInput(sqrtPriceCurrentX96, liquidity, amountRemainingLessFee, zeroForOne)
			if err != nil {
				return
			}
			feeAmount.Set(new(big.Int).Sub(new(big.Int).Abs(amountRemaining), amountIn))
		}
		if zeroForOne {
			amountOut, err = GetAmount1Delta(sqrtPriceNextX96, sqrtPriceCurrentX96, liquidity, false)
			if err != nil {
				return
			}
		} else {
			amountOut, err = GetAmount0Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, false)
			if err != nil {
				return
			}
		}
	} else {
		if zeroForOne {
			amountOut, err = GetAmount1Delta(sqrtPriceTargetX96, sqrtPriceCurrentX96, liquidity, false)
			if err != nil {
				return
			}
		} else {
			amountOut, err = GetAmount0Delta(sqrtPriceCurrentX96, sqrtPriceTargetX96, liquidity, false)
			if err != nil {
				return
			}
		}

		if amountRemaining.Cmp(amountOut) >= 0 {
			// `amountOut` is capped by the target price
			sqrtPriceNextX96.Set(sqrtPriceTargetX96)
		} else {
			// cap the output amount to not exceed the remaining output amount
			amountOut.Set(amountRemaining)
			sqrtPriceNextX96, err = GetNextSqrtPriceFromOutput(sqrtPriceCurrentX96, liquidity, amountOut, zeroForOne)
			if err != nil {
				return
			}
		}

		if zeroForOne {
			amountIn, err = GetAmount0Delta(sqrtPriceNextX96, sqrtPriceCurrentX96, liquidity, true)
			if err != nil {
				return
			}
		} else {
			amountIn, err = GetAmount1Delta(sqrtPriceCurrentX96, sqrtPriceNextX96, liquidity, true)
			if err != nil {
				return
			}
		}

		// `feePips` cannot be `MAX_SWAP_FEE` for exact out
		tmp := new(big.Int)
		denominator := new(big.Int).Sub(MaxSwapFee, big.NewInt(int64(_feePips)))
		tmp, err = MulDivRoundingUp(amountIn, big.NewInt(int64(_feePips)), denominator)
		if err != nil {
			return
		}
		feeAmount.Set(tmp)
	}

	return
}

// GetSqrtPriceTarget computes the next sqrt price for a swap step.
//
// zeroForOne: true if swapping 0→1, false if 1→0
// sqrtPriceNextX96: next initialized tick price (Q64.96)
// sqrtPriceLimitX96: price limit (Q64.96)
func GetSqrtPriceTarget(zeroForOne bool, sqrtPriceNextX96, sqrtPriceLimitX96 *big.Int) *big.Int {
	target := new(big.Int)
	if zeroForOne {
		// 0→1 swap, price cannot go below limit
		if sqrtPriceNextX96.Cmp(sqrtPriceLimitX96) < 0 {
			target.Set(sqrtPriceLimitX96)
		} else {
			target.Set(sqrtPriceNextX96)
		}
	} else {
		// 1→0 swap, price cannot go above limit
		if sqrtPriceNextX96.Cmp(sqrtPriceLimitX96) > 0 {
			target.Set(sqrtPriceLimitX96)
		} else {
			target.Set(sqrtPriceNextX96)
		}
	}
	return target
}
