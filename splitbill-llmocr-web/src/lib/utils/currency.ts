/**
 * Formats a number as currency with thousand separators
 * @param amount - The amount to format
 * @param currency - The currency symbol (default: '$')
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted currency string (e.g., "$1,234.56")
 */
export function formatCurrency(
  amount: number, 
  currency: string = '$', 
  decimals: number = 2
): string {
  // Handle invalid numbers
  if (isNaN(amount) || !isFinite(amount)) {
    return `${currency}0.${'0'.repeat(decimals)}`;
  }

  // Use toLocaleString for proper thousand separators and decimal formatting
  const formatted = amount.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  });

  return `${currency}${formatted}`;
}

/**
 * Formats a number as currency without the currency symbol
 * Useful for calculations or when currency symbol is displayed separately
 * @param amount - The amount to format
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted number string (e.g., "1,234.56")
 */
export function formatAmount(amount: number, decimals: number = 2): string {
  // Handle invalid numbers
  if (isNaN(amount) || !isFinite(amount)) {
    return `0.${'0'.repeat(decimals)}`;
  }

  return amount.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  });
}
