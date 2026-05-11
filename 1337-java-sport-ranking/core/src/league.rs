use chrono::Utc;
use std::collections::{HashMap, HashSet};
use crate::models::{Match, MatchResult, StandingsEntry, ValidationError};

#[derive(Debug, Clone, Default)]
pub struct League {
    matches: Vec<Match>,
    teams: HashSet<String>,
}

impl League {
    pub fn new() -> Self {
        League {
            matches: Vec::new(),
            teams: HashSet::new(),
        }
    }

    pub fn add_match(&mut self, result: MatchResult) -> Result<(), ValidationError> {
        self.validate_match(&result)?;

        let match_record = Match {
            home_team: result.home_team.clone(),
            away_team: result.away_team.clone(),
            home_goals: result.home_goals,
            away_goals: result.away_goals,
            round: result.round,
            date: result.date,
        };

        self.teams.insert(result.home_team);
        self.teams.insert(result.away_team);
        self.matches.push(match_record);

        Ok(())
    }

    fn validate_match(&self, result: &MatchResult) -> Result<(), ValidationError> {
        if result.home_goals < 0 {
            return Err(ValidationError {
                field: "home_goals".to_string(),
                message: "进球数不能为负".to_string(),
            });
        }

        if result.away_goals < 0 {
            return Err(ValidationError {
                field: "away_goals".to_string(),
                message: "进球数不能为负".to_string(),
            });
        }

        if result.home_team == result.away_team {
            return Err(ValidationError {
                field: "teams".to_string(),
                message: "主队和客队不能相同".to_string(),
            });
        }

        let now = Utc::now();
        if result.date > now {
            return Err(ValidationError {
                field: "date".to_string(),
                message: "比赛日期不能是未来".to_string(),
            });
        }

        for existing_match in &self.matches {
            if existing_match.round == result.round {
                if existing_match.home_team == result.home_team
                    || existing_match.away_team == result.home_team
                {
                    return Err(ValidationError {
                        field: "round".to_string(),
                        message: format!("同一轮次 {} 队伍不能出现两次", result.home_team),
                    });
                }

                if existing_match.home_team == result.away_team
                    || existing_match.away_team == result.away_team
                {
                    return Err(ValidationError {
                        field: "round".to_string(),
                        message: format!("同一轮次 {} 队伍不能出现两次", result.away_team),
                    });
                }
            }
        }

        Ok(())
    }

    pub fn get_matches(&self) -> &[Match] {
        &self.matches
    }

    pub fn get_teams(&self) -> &HashSet<String> {
        &self.teams
    }

    pub fn add_team(&mut self, team_name: String) {
        self.teams.insert(team_name);
    }

    pub fn get_standings(&self) -> Vec<StandingsEntry> {
        let mut team_stats: HashMap<String, TeamStats> = HashMap::new();

        for team in &self.teams {
            team_stats.insert(team.clone(), TeamStats::new(team.clone()));
        }

        for match_record in &self.matches {
            Self::update_stats(
                &mut team_stats,
                &match_record.home_team,
                &match_record.away_team,
                match_record.home_goals,
                match_record.away_goals,
            );
            Self::update_stats(
                &mut team_stats,
                &match_record.away_team,
                &match_record.home_team,
                match_record.away_goals,
                match_record.home_goals,
            );
        }

        let mut stats_list: Vec<TeamStats> = team_stats.into_values().collect();

        stats_list.sort_by(|a, b| {
            let a_points = a.points;
            let b_points = b.points;

            if a_points != b_points {
                return b_points.cmp(&a_points);
            }

            let a_gd = a.goals_for - a.goals_against;
            let b_gd = b.goals_for - b.goals_against;
            if a_gd != b_gd {
                return b_gd.cmp(&a_gd);
            }

            if a.goals_for != b.goals_for {
                return b.goals_for.cmp(&a.goals_for);
            }

            let h2h = self.head_to_head(&a.team, &b.team);
            if h2h.is_some() {
                let (a_h2h, b_h2h) = h2h.unwrap();
                if a_h2h.points != b_h2h.points {
                    return b_h2h.points.cmp(&a_h2h.points);
                }
                let a_h2h_gd = a_h2h.goals_for - a_h2h.goals_against;
                let b_h2h_gd = b_h2h.goals_for - b_h2h.goals_against;
                if a_h2h_gd != b_h2h_gd {
                    return b_h2h_gd.cmp(&a_h2h_gd);
                }
            }

            a.team.cmp(&b.team)
        });

        let has_played: Vec<TeamStats> = stats_list
            .iter()
            .filter(|s| s.played > 0)
            .cloned()
            .collect();
        let no_matches: Vec<TeamStats> = stats_list
            .iter()
            .filter(|s| s.played == 0)
            .cloned()
            .collect();

        let mut final_list = has_played;
        let mut sorted_no_matches = no_matches;
        sorted_no_matches.sort_by(|a, b| a.team.cmp(&b.team));
        final_list.extend(sorted_no_matches);

        final_list
            .iter()
            .enumerate()
            .map(|(index, stats)| StandingsEntry {
                rank: index + 1,
                team: stats.team.clone(),
                played: stats.played,
                won: stats.won,
                drawn: stats.drawn,
                lost: stats.lost,
                goals_for: stats.goals_for,
                goals_against: stats.goals_against,
                goal_difference: stats.goals_for - stats.goals_against,
                points: stats.points,
            })
            .collect()
    }

    fn update_stats(
        stats: &mut HashMap<String, TeamStats>,
        team: &str,
        _opponent: &str,
        goals_for: i32,
        goals_against: i32,
    ) {
        if let Some(team_stat) = stats.get_mut(team) {
            team_stat.played += 1;
            team_stat.goals_for += goals_for;
            team_stat.goals_against += goals_against;

            if goals_for > goals_against {
                team_stat.won += 1;
                team_stat.points += 3;
            } else if goals_for == goals_against {
                team_stat.drawn += 1;
                team_stat.points += 1;
            } else {
                team_stat.lost += 1;
            }
        }
    }

    fn head_to_head(&self, team_a: &str, team_b: &str) -> Option<(H2HStats, H2HStats)> {
        let mut a_stats = H2HStats::new();
        let mut b_stats = H2HStats::new();
        let mut has_match = false;

        for match_record in &self.matches {
            let (home, away) = (&match_record.home_team, &match_record.away_team);

            if (home == team_a && away == team_b) || (home == team_b && away == team_a) {
                has_match = true;
                if home == team_a {
                    a_stats.goals_for += match_record.home_goals;
                    a_stats.goals_against += match_record.away_goals;
                    b_stats.goals_for += match_record.away_goals;
                    b_stats.goals_against += match_record.home_goals;

                    if match_record.home_goals > match_record.away_goals {
                        a_stats.points += 3;
                    } else if match_record.home_goals == match_record.away_goals {
                        a_stats.points += 1;
                        b_stats.points += 1;
                    } else {
                        b_stats.points += 3;
                    }
                } else {
                    b_stats.goals_for += match_record.home_goals;
                    b_stats.goals_against += match_record.away_goals;
                    a_stats.goals_for += match_record.away_goals;
                    a_stats.goals_against += match_record.home_goals;

                    if match_record.home_goals > match_record.away_goals {
                        b_stats.points += 3;
                    } else if match_record.home_goals == match_record.away_goals {
                        a_stats.points += 1;
                        b_stats.points += 1;
                    } else {
                        a_stats.points += 3;
                    }
                }
            }
        }

        if has_match {
            Some((a_stats, b_stats))
        } else {
            None
        }
    }
}

#[derive(Debug, Clone)]
struct TeamStats {
    team: String,
    played: i32,
    won: i32,
    drawn: i32,
    lost: i32,
    goals_for: i32,
    goals_against: i32,
    points: i32,
}

impl TeamStats {
    fn new(team: String) -> Self {
        TeamStats {
            team,
            played: 0,
            won: 0,
            drawn: 0,
            lost: 0,
            goals_for: 0,
            goals_against: 0,
            points: 0,
        }
    }
}

#[derive(Debug, Clone)]
struct H2HStats {
    goals_for: i32,
    goals_against: i32,
    points: i32,
}

impl H2HStats {
    fn new() -> Self {
        H2HStats {
            goals_for: 0,
            goals_against: 0,
            points: 0,
        }
    }
}
