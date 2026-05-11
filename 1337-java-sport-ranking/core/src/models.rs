use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct Team {
    pub name: String,
}

impl Team {
    pub fn new(name: impl Into<String>) -> Self {
        Team { name: name.into() }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Match {
    pub home_team: String,
    pub away_team: String,
    pub home_goals: i32,
    pub away_goals: i32,
    pub round: i32,
    pub date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MatchResult {
    pub home_team: String,
    pub away_team: String,
    pub home_goals: i32,
    pub away_goals: i32,
    pub round: i32,
    pub date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StandingsEntry {
    pub rank: usize,
    pub team: String,
    pub played: i32,
    pub won: i32,
    pub drawn: i32,
    pub lost: i32,
    pub goals_for: i32,
    pub goals_against: i32,
    pub goal_difference: i32,
    pub points: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationError {
    pub field: String,
    pub message: String,
}

impl std::fmt::Display for ValidationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}: {}", self.field, self.message)
    }
}

impl std::error::Error for ValidationError {}
