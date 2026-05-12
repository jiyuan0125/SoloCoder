export const calculateMean = (values: number[]): number => {
  if (values.length === 0) return 0;
  return values.reduce((sum, val) => sum + val, 0) / values.length;
};

export const calculateVariance = (values: number[]): number => {
  if (values.length < 2) return 0;
  const mean = calculateMean(values);
  const squaredDiffs = values.map(val => Math.pow(val - mean, 2));
  return squaredDiffs.reduce((sum, val) => sum + val, 0) / (values.length - 1);
};

export const calculateTTest = (
  control: number[],
  treatment: number[]
): { tStat: number; pValue: number } | null => {
  if (control.length < 30 || treatment.length < 30) {
    return null;
  }

  const n1 = control.length;
  const n2 = treatment.length;
  const mean1 = calculateMean(control);
  const mean2 = calculateMean(treatment);
  const var1 = calculateVariance(control);
  const var2 = calculateVariance(treatment);

  const pooledVariance = ((n1 - 1) * var1 + (n2 - 1) * var2) / (n1 + n2 - 2);
  const pooledStd = Math.sqrt(pooledVariance * (1 / n1 + 1 / n2));

  if (pooledStd === 0) {
    return { tStat: 0, pValue: 1 };
  }

  const tStat = (mean2 - mean1) / pooledStd;
  const df = n1 + n2 - 2;
  const pValue = twoTailedPValue(tStat, df);

  return { tStat, pValue };
};

const twoTailedPValue = (tStat: number, df: number): number => {
  const t = Math.abs(tStat);
  
  if (df <= 0) return 1;
  
  if (df >= 100) {
    return 2 * (1 - normalCDF(t));
  }
  
  const t2 = t * t;
  let gamma = 1;
  const halfDf = df / 2;
  
  for (let i = df - 2; i >= 2; i -= 2) {
    gamma = 1 + (i - 1) / (i * (1 + t2 / i)) * gamma;
  }
  
  const denom = Math.sqrt(df * Math.PI) * gamma;
  const numerator = Math.exp(-halfDf * Math.log(1 + t2 / df));
  
  let prob = numerator / denom;
  
  if (t > 0) {
    for (let i = df - 2; i >= 1; i -= 2) {
      prob = prob * i / (i + 1) * (1 + t2 / (i + 1)) + (t > 0 ? 0.5 : 0.5);
      if (i === 1) break;
    }
  }
  
  return 2 * Math.min(0.5, Math.max(0, prob));
};

const normalCDF = (x: number): number => {
  const a1 = 0.254829592;
  const a2 = -0.284496736;
  const a3 = 1.421413741;
  const a4 = -1.453152027;
  const a5 = 1.061405429;
  const p = 0.3275911;

  const sign = x < 0 ? -1 : 1;
  const absX = Math.abs(x);
  const t = 1 / (1 + p * absX);
  const y = 1 - ((((a5 * t + a4) * t + a3) * t + a2) * t + a1) * t * Math.exp(-absX * absX);

  return 0.5 * (1 + sign * y);
};
