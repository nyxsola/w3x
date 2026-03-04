package utils

import "github.com/aicora/go-uniswap/core/libraries"

/**
 * ComputePriceImpact calculates the **price impact** of a trade relative to the mid price.
 *
 * Price impact measures how much the execution price deviates from the market mid price
 * due to trade size and AMM liquidity effects.
 *
 * Mathematically:
 *
 *     priceImpact = (expectedOutput - actualOutput) / expectedOutput
 *
 * Where:
 *   - expectedOutput is the amount of output token that would be received
 *     if the trade occurred at the midPrice.
 *   - actualOutput is the amount of output token actually received (or expected
 *     from the AMM pool).
 *
 * This function **does not predict output amounts**; it requires `outputAmount`
 * to be provided (e.g., computed from the AMM formula or route aggregator).
 *
 * Parameters:
 *   - midPrice: the mid-market price of the pair before the trade
 *               (Price = Token1 / Token0)
 *   - inputAmount: the input amount of the token0 for this trade
 *   - outputAmount: the actual or expected output amount of the token1
 *
 * Returns:
 *   - Percent representing the price impact of the trade
 *   - error if the currencies do not match or the quote fails
 *
 * Usage example:
 *
 *     eth := NewCurrency(1, "0xETH...", 18, "ETH", "Ether")
 *     usdc := NewCurrency(1, "0xUSDC...", 6, "USDC", "US Dollar Coin")
 *
 *     midPrice := NewPrice(eth, usdc, big.NewInt(1e18), big.NewInt(2000e6))
 *     input := NewCurrencyAmount(eth, big.NewInt(1e18), big.NewInt(1))
 *     output := NewCurrencyAmount(usdc, big.NewInt(1990e6), big.NewInt(1))
 *
 *     impact, err := ComputePriceImpact(midPrice, input, output)
 *     // impact will be ~0.5%
 *
 * Notes:
 *   - Always use the same token pair for midPrice, inputAmount, and outputAmount.
 *   - For on-chain safety, ensure that outputAmount is consistent with the AMM
 *     formula or expected pool reserves.
 *   - Price impact is typically used to display slippage risk to users or
 *     to optimize routing.
 */
func ComputePriceImpact(midPrice *libraries.Price, inputAmount, outputAmount *libraries.CurrencyAmount) (*libraries.Percent, error) {
	token1Amount, err := midPrice.Quote(inputAmount)
	if err != nil {
		return nil, err
	}
	priceImpact := token1Amount.Subtract(outputAmount).Divide(token1Amount.Fraction)
	return libraries.NewPercent(priceImpact.Numerator, priceImpact.Denominator), nil
}
