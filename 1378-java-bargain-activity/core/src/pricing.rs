use rand::Rng;

pub fn calculate_bargain_amount(
    current_price: f64,
    floor_price: f64,
    bargain_count: u32,
) -> f64 {
    let mut rng = rand::thread_rng();
    
    let range = match bargain_count {
        0 => (0.05, 0.10),
        1 => (0.03, 0.07),
        2 => (0.01, 0.04),
        _ => (0.005, 0.02),
    };
    
    let percentage = rng.gen_range(range.0..range.1);
    let raw_amount = current_price * percentage;
    
    let max_possible = current_price - floor_price;
    if max_possible <= 0.0 {
        0.0
    } else {
        raw_amount.min(max_possible)
    }
}

pub fn round_price(price: f64) -> f64 {
    (price * 100.0).round() / 100.0
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_bargain_amount_decreases() {
        let original = 100.0;
        let floor = 50.0;
        
        let amounts: Vec<f64> = (0..5).map(|i| {
            calculate_bargain_amount(original, floor, i)
        }).collect();
        
        assert!(amounts[0] >= 5.0 && amounts[0] <= 10.0);
        assert!(amounts[1] >= 3.0 && amounts[1] <= 7.0);
        assert!(amounts[2] >= 1.0 && amounts[2] <= 4.0);
        for &a in &amounts[3..] {
            assert!(a >= 0.5 && a <= 2.0);
        }
    }
    
    #[test]
    fn test_bargain_not_below_floor() {
        let current = 51.0;
        let floor = 50.0;
        
        let amount = calculate_bargain_amount(current, floor, 0);
        assert!(amount <= 1.0);
    }
    
    #[test]
    fn test_bargain_at_floor_returns_zero() {
        let current = 50.0;
        let floor = 50.0;
        
        let amount = calculate_bargain_amount(current, floor, 0);
        assert_eq!(amount, 0.0);
    }
}
