use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum AqlLevel {
    Level065,
    Level10,
    Level15,
    Level25,
    Level40,
    Level65,
}

impl AqlLevel {
    pub fn as_str(&self) -> &'static str {
        match self {
            AqlLevel::Level065 => "0.65",
            AqlLevel::Level10 => "1.0",
            AqlLevel::Level15 => "1.5",
            AqlLevel::Level25 => "2.5",
            AqlLevel::Level40 => "4.0",
            AqlLevel::Level65 => "6.5",
        }
    }

    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "0.65" => Some(AqlLevel::Level065),
            "1.0" => Some(AqlLevel::Level10),
            "1.5" => Some(AqlLevel::Level15),
            "2.5" => Some(AqlLevel::Level25),
            "4.0" => Some(AqlLevel::Level40),
            "6.5" => Some(AqlLevel::Level65),
            _ => None,
        }
    }
}

#[derive(Clone)]
pub struct AqlConfig {
    pub sample_size_table: Vec<(u32, u32, u32)>,
    pub ac_table: HashMap<u32, HashMap<AqlLevel, u32>>,
}

impl AqlConfig {
    pub fn new() -> Self {
        let sample_size_table = vec![
            (2, 8, 2),
            (9, 15, 3),
            (16, 25, 5),
            (26, 50, 8),
            (51, 90, 13),
            (91, 150, 20),
            (151, 280, 32),
            (281, 500, 50),
            (501, 1200, 80),
            (1201, 3200, 125),
            (3201, 10000, 200),
            (10001, 35000, 315),
            (35001, 500000, 500),
            (500001, u32::MAX, 800),
        ];

        let mut ac_table = HashMap::new();

        let mut ac_2 = HashMap::new();
        ac_2.insert(AqlLevel::Level065, 0);
        ac_2.insert(AqlLevel::Level10, 0);
        ac_2.insert(AqlLevel::Level15, 0);
        ac_2.insert(AqlLevel::Level25, 0);
        ac_2.insert(AqlLevel::Level40, 0);
        ac_2.insert(AqlLevel::Level65, 1);
        ac_table.insert(2, ac_2);

        let mut ac_3 = HashMap::new();
        ac_3.insert(AqlLevel::Level065, 0);
        ac_3.insert(AqlLevel::Level10, 0);
        ac_3.insert(AqlLevel::Level15, 0);
        ac_3.insert(AqlLevel::Level25, 0);
        ac_3.insert(AqlLevel::Level40, 1);
        ac_3.insert(AqlLevel::Level65, 1);
        ac_table.insert(3, ac_3);

        let mut ac_5 = HashMap::new();
        ac_5.insert(AqlLevel::Level065, 0);
        ac_5.insert(AqlLevel::Level10, 0);
        ac_5.insert(AqlLevel::Level15, 0);
        ac_5.insert(AqlLevel::Level25, 1);
        ac_5.insert(AqlLevel::Level40, 1);
        ac_5.insert(AqlLevel::Level65, 2);
        ac_table.insert(5, ac_5);

        let mut ac_8 = HashMap::new();
        ac_8.insert(AqlLevel::Level065, 0);
        ac_8.insert(AqlLevel::Level10, 0);
        ac_8.insert(AqlLevel::Level15, 1);
        ac_8.insert(AqlLevel::Level25, 1);
        ac_8.insert(AqlLevel::Level40, 2);
        ac_8.insert(AqlLevel::Level65, 3);
        ac_table.insert(8, ac_8);

        let mut ac_13 = HashMap::new();
        ac_13.insert(AqlLevel::Level065, 0);
        ac_13.insert(AqlLevel::Level10, 1);
        ac_13.insert(AqlLevel::Level15, 1);
        ac_13.insert(AqlLevel::Level25, 2);
        ac_13.insert(AqlLevel::Level40, 3);
        ac_13.insert(AqlLevel::Level65, 5);
        ac_table.insert(13, ac_13);

        let mut ac_20 = HashMap::new();
        ac_20.insert(AqlLevel::Level065, 1);
        ac_20.insert(AqlLevel::Level10, 1);
        ac_20.insert(AqlLevel::Level15, 2);
        ac_20.insert(AqlLevel::Level25, 3);
        ac_20.insert(AqlLevel::Level40, 5);
        ac_20.insert(AqlLevel::Level65, 7);
        ac_table.insert(20, ac_20);

        let mut ac_32 = HashMap::new();
        ac_32.insert(AqlLevel::Level065, 1);
        ac_32.insert(AqlLevel::Level10, 2);
        ac_32.insert(AqlLevel::Level15, 3);
        ac_32.insert(AqlLevel::Level25, 5);
        ac_32.insert(AqlLevel::Level40, 7);
        ac_32.insert(AqlLevel::Level65, 10);
        ac_table.insert(32, ac_32);

        let mut ac_50 = HashMap::new();
        ac_50.insert(AqlLevel::Level065, 2);
        ac_50.insert(AqlLevel::Level10, 3);
        ac_50.insert(AqlLevel::Level15, 5);
        ac_50.insert(AqlLevel::Level25, 7);
        ac_50.insert(AqlLevel::Level40, 10);
        ac_50.insert(AqlLevel::Level65, 14);
        ac_table.insert(50, ac_50);

        let mut ac_80 = HashMap::new();
        ac_80.insert(AqlLevel::Level065, 3);
        ac_80.insert(AqlLevel::Level10, 5);
        ac_80.insert(AqlLevel::Level15, 7);
        ac_80.insert(AqlLevel::Level25, 10);
        ac_80.insert(AqlLevel::Level40, 14);
        ac_80.insert(AqlLevel::Level65, 21);
        ac_table.insert(80, ac_80);

        let mut ac_125 = HashMap::new();
        ac_125.insert(AqlLevel::Level065, 5);
        ac_125.insert(AqlLevel::Level10, 7);
        ac_125.insert(AqlLevel::Level15, 10);
        ac_125.insert(AqlLevel::Level25, 14);
        ac_125.insert(AqlLevel::Level40, 21);
        ac_125.insert(AqlLevel::Level65, 21);
        ac_table.insert(125, ac_125);

        let mut ac_200 = HashMap::new();
        ac_200.insert(AqlLevel::Level065, 7);
        ac_200.insert(AqlLevel::Level10, 10);
        ac_200.insert(AqlLevel::Level15, 14);
        ac_200.insert(AqlLevel::Level25, 21);
        ac_200.insert(AqlLevel::Level40, 21);
        ac_200.insert(AqlLevel::Level65, 21);
        ac_table.insert(200, ac_200);

        let mut ac_315 = HashMap::new();
        ac_315.insert(AqlLevel::Level065, 10);
        ac_315.insert(AqlLevel::Level10, 14);
        ac_315.insert(AqlLevel::Level15, 21);
        ac_315.insert(AqlLevel::Level25, 21);
        ac_315.insert(AqlLevel::Level40, 21);
        ac_315.insert(AqlLevel::Level65, 21);
        ac_table.insert(315, ac_315);

        let mut ac_500 = HashMap::new();
        ac_500.insert(AqlLevel::Level065, 14);
        ac_500.insert(AqlLevel::Level10, 21);
        ac_500.insert(AqlLevel::Level15, 21);
        ac_500.insert(AqlLevel::Level25, 21);
        ac_500.insert(AqlLevel::Level40, 21);
        ac_500.insert(AqlLevel::Level65, 21);
        ac_table.insert(500, ac_500);

        let mut ac_800 = HashMap::new();
        ac_800.insert(AqlLevel::Level065, 21);
        ac_800.insert(AqlLevel::Level10, 21);
        ac_800.insert(AqlLevel::Level15, 21);
        ac_800.insert(AqlLevel::Level25, 21);
        ac_800.insert(AqlLevel::Level40, 21);
        ac_800.insert(AqlLevel::Level65, 21);
        ac_table.insert(800, ac_800);

        AqlConfig {
            sample_size_table,
            ac_table,
        }
    }

    pub fn get_sample_size(&self, batch_size: u32) -> u32 {
        for &(min, max, sample) in &self.sample_size_table {
            if batch_size >= min && batch_size <= max {
                return sample;
            }
        }
        2
    }

    pub fn get_ac(&self, sample_size: u32, aql_level: AqlLevel) -> u32 {
        self.ac_table
            .get(&sample_size)
            .and_then(|level_map| level_map.get(&aql_level))
            .copied()
            .unwrap_or(0)
    }
}

impl Default for AqlConfig {
    fn default() -> Self {
        Self::new()
    }
}
