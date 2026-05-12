export function yuanToCents(yuan: number): number {
  return Math.round(yuan * 100);
}

export function centsToYuan(cents: number): string {
  return (cents / 100).toFixed(2);
}

export function formatMoneyForResponse(cents: number): string {
  return centsToYuan(cents);
}
