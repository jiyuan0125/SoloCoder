use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum PerformanceLevel {
    P1 = 1,
    P2 = 2,
    P3 = 3,
    P4 = 4,
    P5 = 5,
}

impl PerformanceLevel {
    pub fn from_value(value: u8) -> Option<Self> {
        match value {
            1 => Some(Self::P1),
            2 => Some(Self::P2),
            3 => Some(Self::P3),
            4 => Some(Self::P4),
            5 => Some(Self::P5),
            _ => None,
        }
    }
    
    pub fn value(&self) -> u8 {
        *self as u8
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum PotentialLevel {
    L1 = 1,
    L2 = 2,
    L3 = 3,
    L4 = 4,
    L5 = 5,
}

impl PotentialLevel {
    pub fn from_value(value: u8) -> Option<Self> {
        match value {
            1 => Some(Self::L1),
            2 => Some(Self::L2),
            3 => Some(Self::L3),
            4 => Some(Self::L4),
            5 => Some(Self::L5),
            _ => None,
        }
    }
    
    pub fn value(&self) -> u8 {
        *self as u8
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum GridZone {
    A1,
    A2,
    A3,
    B1,
    B2,
    B3,
    C1,
    C2,
    C3,
}

impl GridZone {
    pub fn calculate(performance: PerformanceLevel, potential: PotentialLevel) -> Self {
        let p = performance.value();
        let l = potential.value();
        
        if p <= 2 {
            if l <= 2 {
                GridZone::C3
            } else if l <= 3 {
                GridZone::C2
            } else {
                GridZone::C1
            }
        } else if p <= 3 {
            if l <= 2 {
                GridZone::B3
            } else if l <= 3 {
                GridZone::B2
            } else {
                GridZone::B1
            }
        } else {
            if l <= 2 {
                GridZone::A3
            } else if l <= 3 {
                GridZone::A2
            } else {
                GridZone::A1
            }
        }
    }
    
    pub fn name(&self) -> &'static str {
        match self {
            GridZone::A1 => "明星员工",
            GridZone::A2 => "高绩效稳定型",
            GridZone::A3 => "绩效贡献者",
            GridZone::B1 => "高潜力成长型",
            GridZone::B2 => "中坚力量",
            GridZone::B3 => "待发展稳定型",
            GridZone::C1 => "待观察高潜型",
            GridZone::C2 => "待改进型",
            GridZone::C3 => "需优化型",
        }
    }
    
    pub fn description(&self) -> &'static str {
        match self {
            GridZone::A1 => "高绩效、高潜力。是组织的未来领导者，需要重点培养和保留。",
            GridZone::A2 => "高绩效、中等潜力。是组织的核心骨干，应给予认可和适当的发展机会。",
            GridZone::A3 => "高绩效、低潜力。是组织的稳定贡献者，应给予认可并考虑横向发展。",
            GridZone::B1 => "中等绩效、高潜力。需要重点关注和培养，提升绩效表现。",
            GridZone::B2 => "中等绩效、中等潜力。是组织的基础力量，需要持续关注和发展。",
            GridZone::B3 => "中等绩效、低潜力。需要明确发展路径，考虑转岗或提升技能。",
            GridZone::C1 => "低绩效、高潜力。需要密切关注，找出绩效问题并提供支持。",
            GridZone::C2 => "低绩效、中等潜力。需要制定改进计划，设定明确的绩效目标。",
            GridZone::C3 => "低绩效、低潜力。需要认真评估，考虑转岗或优化。",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Department {
    pub id: Uuid,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

impl Department {
    pub fn new(name: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: Uuid,
    pub name: String,
    pub employee_number: String,
    pub department_id: Uuid,
    pub position: String,
    pub created_at: DateTime<Utc>,
}

impl Employee {
    pub fn new(name: String, employee_number: String, department_id: Uuid, position: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            employee_number,
            department_id,
            position,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Assessment {
    pub id: Uuid,
    pub employee_id: Uuid,
    pub year: i32,
    pub performance: PerformanceLevel,
    pub potential: PotentialLevel,
    pub zone: GridZone,
    pub created_at: DateTime<Utc>,
}

impl Assessment {
    pub fn new(employee_id: Uuid, year: i32, performance: PerformanceLevel, potential: PotentialLevel) -> Self {
        let zone = GridZone::calculate(performance, potential);
        Self {
            id: Uuid::new_v4(),
            employee_id,
            year,
            performance,
            potential,
            zone,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SuccessionPlan {
    pub id: Uuid,
    pub position: String,
    pub department_id: Uuid,
    pub is_key_position: bool,
    pub candidates: Vec<Candidate>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl SuccessionPlan {
    pub fn new(position: String, department_id: Uuid, is_key_position: bool) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            position,
            department_id,
            is_key_position,
            candidates: Vec::new(),
            created_at: now,
            updated_at: now,
        }
    }
    
    pub fn has_enough_candidates(&self) -> bool {
        !self.is_key_position || self.candidates.len() >= 2
    }
    
    pub fn warning_message(&self) -> Option<String> {
        if self.is_key_position && self.candidates.len() < 2 {
            Some(format!(
                "关键岗位「{}」的继任候选人不足2人，当前仅有{}人",
                self.position,
                self.candidates.len()
            ))
        } else {
            None
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Candidate {
    pub employee_id: Uuid,
    pub current_zone: GridZone,
    pub assessment_year: i32,
    pub notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GridDistribution {
    pub year: i32,
    pub department_id: Option<Uuid>,
    pub counts: GridCounts,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct GridCounts {
    pub a1: usize,
    pub a2: usize,
    pub a3: usize,
    pub b1: usize,
    pub b2: usize,
    pub b3: usize,
    pub c1: usize,
    pub c2: usize,
    pub c3: usize,
    pub total: usize,
}

impl GridCounts {
    pub fn increment(&mut self, zone: GridZone) {
        match zone {
            GridZone::A1 => self.a1 += 1,
            GridZone::A2 => self.a2 += 1,
            GridZone::A3 => self.a3 += 1,
            GridZone::B1 => self.b1 += 1,
            GridZone::B2 => self.b2 += 1,
            GridZone::B3 => self.b3 += 1,
            GridZone::C1 => self.c1 += 1,
            GridZone::C2 => self.c2 += 1,
            GridZone::C3 => self.c3 += 1,
        }
        self.total += 1;
    }
}
