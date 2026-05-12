use crate::types::MetricValue;

#[derive(Debug, Clone)]
pub struct SummaryWindow {
    window_size: usize,
    samples: Vec<f64>,
    next_idx: usize,
    count: usize,
    sum: f64,
    min: f64,
    max: f64,
}

impl SummaryWindow {
    pub fn new(window_size: usize) -> Self {
        Self {
            window_size,
            samples: Vec::with_capacity(window_size),
            next_idx: 0,
            count: 0,
            sum: 0.0,
            min: f64::INFINITY,
            max: f64::NEG_INFINITY,
        }
    }

    pub fn push(&mut self, value: f64) {
        self.sum += value;

        if self.count < self.window_size {
            self.samples.push(value);
            self.count += 1;
        } else {
            let old_value = self.samples[self.next_idx];
            self.sum -= old_value;
            self.samples[self.next_idx] = value;
        }

        if value < self.min {
            self.min = value;
        }
        if value > self.max {
            self.max = value;
        }

        self.next_idx = (self.next_idx + 1) % self.window_size;
    }

    fn min_safe(&self) -> f64 {
        if self.count == 0 {
            0.0
        } else {
            self.min
        }
    }

    fn max_safe(&self) -> f64 {
        if self.count == 0 {
            0.0
        } else {
            self.max
        }
    }

    fn avg_safe(&self) -> f64 {
        if self.count == 0 {
            0.0
        } else {
            self.sum / self.count as f64
        }
    }

    fn approximate_quantile(&self, p: f64) -> Option<f64> {
        if self.count < 10 {
            return None;
        }

        let n = self.samples.len();
        if n == 0 {
            return None;
        }

        let num_buckets = 20usize.min(n);
        let bucket_size = n / num_buckets;

        let mut bucket_mins: Vec<f64> = Vec::with_capacity(num_buckets);
        let mut bucket_maxs: Vec<f64> = Vec::with_capacity(num_buckets);

        for i in 0..num_buckets {
            let start = i * bucket_size;
            let end = if i == num_buckets - 1 {
                n
            } else {
                (i + 1) * bucket_size
            };

            let slice = &self.samples[start..end];
            let mut local_min = f64::INFINITY;
            let mut local_max = f64::NEG_INFINITY;

            for &val in slice {
                if val < local_min {
                    local_min = val;
                }
                if val > local_max {
                    local_max = val;
                }
            }

            bucket_mins.push(local_min);
            bucket_maxs.push(local_max);
        }

        bucket_mins.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));
        bucket_maxs.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));

        let idx = (p * (num_buckets as f64 - 1.0)) as usize;
        let idx = idx.min(num_buckets - 1);

        let min_val = bucket_mins[idx];
        let max_val = bucket_maxs[idx];

        Some((min_val + max_val) / 2.0)
    }

    pub fn snapshot_value(&self) -> MetricValue {
        if self.count < 10 {
            MetricValue::Summary {
                count: self.count,
                min: self.min_safe(),
                max: self.max_safe(),
                avg: self.avg_safe(),
                p50: None,
                p90: None,
                p99: None,
                error: Some("insufficient data"),
            }
        } else {
            MetricValue::Summary {
                count: self.count,
                min: self.min_safe(),
                max: self.max_safe(),
                avg: self.avg_safe(),
                p50: self.approximate_quantile(0.50),
                p90: self.approximate_quantile(0.90),
                p99: self.approximate_quantile(0.99),
                error: None,
            }
        }
    }
}
