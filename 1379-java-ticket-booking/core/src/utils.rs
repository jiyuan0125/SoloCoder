use chrono::{Datelike, NaiveDate};
use crate::errors::{Result, TicketError};
use crate::models::{Season, ChildTicketCategory};

pub fn get_season(date: &NaiveDate) -> Season {
    match date.month() {
        4..=10 => Season::PeakSeason,
        _ => Season::OffSeason,
    }
}

pub fn validate_id_card(id_card: &str) -> Result<()> {
    if id_card.len() != 18 {
        return Err(TicketError::InvalidIdCard);
    }
    
    let chars: Vec<char> = id_card.chars().collect();
    for c in &chars[0..17] {
        if !c.is_ascii_digit() {
            return Err(TicketError::InvalidIdCard);
        }
    }
    
    let last_char = chars[17];
    if !last_char.is_ascii_digit() && last_char != 'X' && last_char != 'x' {
        return Err(TicketError::InvalidIdCard);
    }
    
    Ok(())
}

pub fn extract_birth_date_from_id_card(id_card: &str) -> Result<NaiveDate> {
    if id_card.len() < 14 {
        return Err(TicketError::InvalidIdCard);
    }
    
    let year_str = &id_card[6..10];
    let month_str = &id_card[10..12];
    let day_str = &id_card[12..14];
    
    let year: i32 = year_str.parse().map_err(|_| TicketError::InvalidIdCard)?;
    let month: u32 = month_str.parse().map_err(|_| TicketError::InvalidIdCard)?;
    let day: u32 = day_str.parse().map_err(|_| TicketError::InvalidIdCard)?;
    
    NaiveDate::from_ymd_opt(year, month, day)
        .ok_or(TicketError::InvalidIdCard)
}

pub fn is_elder(birth_date: &NaiveDate, use_date: &NaiveDate) -> bool {
    let age = use_date.years_since(*birth_date).unwrap_or(0);
    age >= 65
}

pub fn get_child_category(height: f32) -> ChildTicketCategory {
    if height < 1.2 {
        ChildTicketCategory::Free
    } else if height < 1.4 {
        ChildTicketCategory::HalfPrice
    } else {
        ChildTicketCategory::FullPrice
    }
}

pub fn calculate_actual_price(
    base_price: u32,
    season: Season,
    is_half_price: bool,
    is_free: bool,
) -> u32 {
    if is_free {
        return 0;
    }
    
    let season_multiplier = match season {
        Season::PeakSeason => 1.3,
        Season::OffSeason => 0.5,
    };
    
    let mut price = (base_price as f32) * season_multiplier;
    
    if is_half_price {
        price *= 0.5;
    }
    
    price.round() as u32
}

pub fn calculate_refund_fee(total_price: u32) -> u32 {
    let fee = (total_price as f32 * 0.1).round() as u32;
    if fee < 1 {
        1
    } else {
        fee
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::NaiveDate;

    #[test]
    fn test_get_season() {
        let peak_date = NaiveDate::from_ymd_opt(2024, 7, 15).unwrap();
        assert_eq!(get_season(&peak_date), Season::PeakSeason);
        
        let off_date = NaiveDate::from_ymd_opt(2024, 12, 25).unwrap();
        assert_eq!(get_season(&off_date), Season::OffSeason);
        
        let march = NaiveDate::from_ymd_opt(2024, 3, 31).unwrap();
        assert_eq!(get_season(&march), Season::OffSeason);
        
        let april = NaiveDate::from_ymd_opt(2024, 4, 1).unwrap();
        assert_eq!(get_season(&april), Season::PeakSeason);
    }

    #[test]
    fn test_validate_id_card() {
        assert!(validate_id_card("110101199003077654").is_ok());
        assert!(validate_id_card("11010119900307765X").is_ok());
        assert!(validate_id_card("11010119900307765").is_err());
        assert!(validate_id_card("110101199003077654321").is_err());
    }

    #[test]
    fn test_extract_birth_date() {
        let id_card = "110101199003077654";
        let birth_date = extract_birth_date_from_id_card(id_card).unwrap();
        assert_eq!(birth_date, NaiveDate::from_ymd_opt(1990, 3, 7).unwrap());
    }

    #[test]
    fn test_is_elder() {
        let birth = NaiveDate::from_ymd_opt(1955, 1, 1).unwrap();
        let use_date = NaiveDate::from_ymd_opt(2024, 1, 1).unwrap();
        assert!(is_elder(&birth, &use_date));
        
        let young = NaiveDate::from_ymd_opt(1990, 1, 1).unwrap();
        assert!(!is_elder(&young, &use_date));
    }

    #[test]
    fn test_child_category() {
        assert_eq!(get_child_category(1.0), ChildTicketCategory::Free);
        assert_eq!(get_child_category(1.2), ChildTicketCategory::HalfPrice);
        assert_eq!(get_child_category(1.3), ChildTicketCategory::HalfPrice);
        assert_eq!(get_child_category(1.4), ChildTicketCategory::FullPrice);
        assert_eq!(get_child_category(1.5), ChildTicketCategory::FullPrice);
    }

    #[test]
    fn test_calculate_price() {
        let base = 100;
        
        assert_eq!(calculate_actual_price(base, Season::PeakSeason, false, false), 130);
        assert_eq!(calculate_actual_price(base, Season::OffSeason, false, false), 50);
        assert_eq!(calculate_actual_price(base, Season::PeakSeason, true, false), 65);
        assert_eq!(calculate_actual_price(base, Season::OffSeason, true, false), 25);
        assert_eq!(calculate_actual_price(base, Season::PeakSeason, false, true), 0);
    }

    #[test]
    fn test_refund_fee() {
        assert_eq!(calculate_refund_fee(100), 10);
        assert_eq!(calculate_refund_fee(5), 1);
        assert_eq!(calculate_refund_fee(15), 2);
    }
}
