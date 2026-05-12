export const calculateP99 = (data: number[]): number => {
  if (data.length === 0) {
    return 0;
  }

  const sorted = [...data].sort((a, b) => a - b);
  
  const percentile = 99;
  const n = sorted.length;
  const index = (percentile / 100) * (n - 1);
  
  if (Number.isInteger(index)) {
    return sorted[index];
  }
  
  const lower = Math.floor(index);
  const upper = lower + 1;
  const weight = index - lower;
  
  return sorted[lower] + (sorted[upper] - sorted[lower]) * weight;
};
