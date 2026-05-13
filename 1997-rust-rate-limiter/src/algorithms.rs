use crate::types::{
    CounterState, FixedWindowCounter, RateLimitCheckResult, RateLimitRule, SlidingWindowCounter,
    TokenBucketState,
};
use std::time::{Duration, Instant, SystemTime};

pub fn check_fixed_window(
    rule: &RateLimitRule,
    counter: &mut FixedWindowCounter,
    now: Instant,
) -> RateLimitCheckResult {
    let window_duration = Duration::from_secs(rule.window_secs);

    if now.duration_since(counter.window_start) >= window_duration {
        counter.count = 0;
        counter.window_start = now;
    }

    if counter.count >= rule.limit {
        let elapsed = now.duration_since(counter.window_start);
        let remaining = window_duration.saturating_sub(elapsed);
        RateLimitCheckResult {
            allowed: false,
            remaining: 0,
            retry_after: Some(remaining),
        }
    } else {
        counter.count += 1;
        RateLimitCheckResult {
            allowed: true,
            remaining: rule.limit.saturating_sub(counter.count),
            retry_after: None,
        }
    }
}

pub fn check_sliding_window(
    rule: &RateLimitRule,
    counter: &mut SlidingWindowCounter,
    _now: Instant,
) -> RateLimitCheckResult {
    let _window_duration = Duration::from_secs(rule.window_secs);
    let current_sec = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap_or(Duration::ZERO)
        .as_secs();
    let cutoff_sec = current_sec.saturating_sub(rule.window_secs);

    let mut to_remove = Vec::new();
    for (&sec, &count) in counter.requests.iter() {
        if sec <= cutoff_sec {
            to_remove.push(sec);
            counter.total = counter.total.saturating_sub(count);
        } else {
            break;
        }
    }
    for sec in to_remove {
        counter.requests.remove(&sec);
    }

    if counter.total >= rule.limit {
        let oldest_sec = counter.requests.keys().next().copied().unwrap_or(current_sec);
        let wait_secs = oldest_sec.saturating_add(rule.window_secs + 1).saturating_sub(current_sec);
        RateLimitCheckResult {
            allowed: false,
            remaining: 0,
            retry_after: Some(Duration::from_secs(wait_secs)),
        }
    } else {
        *counter.requests.entry(current_sec).or_insert(0) += 1;
        counter.total += 1;
        RateLimitCheckResult {
            allowed: true,
            remaining: rule.limit.saturating_sub(counter.total),
            retry_after: None,
        }
    }
}

pub fn check_token_bucket(
    rule: &RateLimitRule,
    bucket: &mut TokenBucketState,
    now: Instant,
) -> RateLimitCheckResult {
    let elapsed = now.duration_since(bucket.last_refill);
    let intervals = elapsed.as_millis() as u64 / rule.refill_interval.as_millis() as u64;

    if intervals > 0 {
        let tokens_to_add = intervals.saturating_mul(rule.refill_rate);
        bucket.tokens = std::cmp::min(bucket.tokens.saturating_add(tokens_to_add), rule.capacity);
        bucket.last_refill = bucket
            .last_refill
            .checked_add(Duration::from_millis(intervals * rule.refill_interval.as_millis() as u64))
            .unwrap_or(now);
    }

    if bucket.tokens == 0 {
        let next_refill = bucket
            .last_refill
            .checked_add(rule.refill_interval)
            .unwrap_or(now);
        let wait_time = next_refill.saturating_duration_since(now);
        RateLimitCheckResult {
            allowed: false,
            remaining: 0,
            retry_after: Some(wait_time),
        }
    } else {
        bucket.tokens -= 1;
        RateLimitCheckResult {
            allowed: true,
            remaining: bucket.tokens,
            retry_after: None,
        }
    }
}

pub fn init_counter_for_rule(rule: &RateLimitRule, now: Instant) -> CounterState {
    match rule.algorithm {
        crate::types::AlgorithmType::FixedWindow => {
            CounterState::FixedWindow(FixedWindowCounter {
                count: 0,
                window_start: now,
            })
        }
        crate::types::AlgorithmType::SlidingWindow => {
            CounterState::SlidingWindow(SlidingWindowCounter {
                requests: std::collections::BTreeMap::new(),
                total: 0,
            })
        }
        crate::types::AlgorithmType::TokenBucket => CounterState::TokenBucket(TokenBucketState {
            tokens: rule.capacity,
            last_refill: now,
        }),
    }
}

pub fn check_counter(
    rule: &RateLimitRule,
    counter: &mut CounterState,
    now: Instant,
) -> RateLimitCheckResult {
    match counter {
        CounterState::FixedWindow(c) => check_fixed_window(rule, c, now),
        CounterState::SlidingWindow(c) => check_sliding_window(rule, c, now),
        CounterState::TokenBucket(c) => check_token_bucket(rule, c, now),
    }
}
