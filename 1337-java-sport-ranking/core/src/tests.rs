#[cfg(test)]
mod tests {
    use crate::league::League;
    use crate::models::MatchResult;
    use chrono::{TimeZone, Utc};

    #[test]
    fn test_basic_ranking() {
        let mut league = League::new();

        let date = Utc.with_ymd_and_hms(2024, 1, 15, 14, 0, 0).unwrap();

        league
            .add_match(MatchResult {
                home_team: "A".to_string(),
                away_team: "B".to_string(),
                home_goals: 2,
                away_goals: 1,
                round: 1,
                date,
            })
            .unwrap();

        league
            .add_match(MatchResult {
                home_team: "C".to_string(),
                away_team: "D".to_string(),
                home_goals: 1,
                away_goals: 1,
                round: 1,
                date,
            })
            .unwrap();

        let standings = league.get_standings();

        assert_eq!(standings[0].team, "A");
        assert_eq!(standings[0].points, 3);

        assert_eq!(standings[1].points, 1);
        assert!(standings[1].team == "C" || standings[1].team == "D");

        assert_eq!(standings[3].team, "B");
        assert_eq!(standings[3].points, 0);
    }

    #[test]
    fn test_goal_difference_tiebreaker() {
        let mut league = League::new();
        let date = Utc.with_ymd_and_hms(2024, 1, 15, 14, 0, 0).unwrap();

        league
            .add_match(MatchResult {
                home_team: "A".to_string(),
                away_team: "B".to_string(),
                home_goals: 3,
                away_goals: 0,
                round: 1,
                date,
            })
            .unwrap();

        league
            .add_match(MatchResult {
                home_team: "C".to_string(),
                away_team: "D".to_string(),
                home_goals: 2,
                away_goals: 0,
                round: 2,
                date,
            })
            .unwrap();

        let standings = league.get_standings();

        assert_eq!(standings[0].team, "A");
        assert_eq!(standings[0].goal_difference, 3);
        assert_eq!(standings[1].team, "C");
        assert_eq!(standings[1].goal_difference, 2);
    }

    #[test]
    fn test_validation_negative_goals() {
        let mut league = League::new();
        let date = Utc.with_ymd_and_hms(2024, 1, 15, 14, 0, 0).unwrap();

        let result = league.add_match(MatchResult {
            home_team: "A".to_string(),
            away_team: "B".to_string(),
            home_goals: -1,
            away_goals: 1,
            round: 1,
            date,
        });

        assert!(result.is_err());
    }

    #[test]
    fn test_validation_same_team_in_round() {
        let mut league = League::new();
        let date = Utc.with_ymd_and_hms(2024, 1, 15, 14, 0, 0).unwrap();

        league
            .add_match(MatchResult {
                home_team: "A".to_string(),
                away_team: "B".to_string(),
                home_goals: 1,
                away_goals: 1,
                round: 1,
                date,
            })
            .unwrap();

        let result = league.add_match(MatchResult {
            home_team: "A".to_string(),
            away_team: "C".to_string(),
            home_goals: 2,
            away_goals: 0,
            round: 1,
            date,
        });

        assert!(result.is_err());
    }

    #[test]
    fn test_validation_future_date() {
        let mut league = League::new();
        let future_date = Utc::now() + chrono::Duration::days(1);

        let result = league.add_match(MatchResult {
            home_team: "A".to_string(),
            away_team: "B".to_string(),
            home_goals: 1,
            away_goals: 0,
            round: 1,
            date: future_date,
        });

        assert!(result.is_err());
    }

    #[test]
    fn test_teams_without_matches_at_end() {
        let mut league = League::new();
        league.add_team("X".to_string());
        league.add_team("Y".to_string());
        league.add_team("Z".to_string());

        let date = Utc.with_ymd_and_hms(2024, 1, 15, 14, 0, 0).unwrap();
        league
            .add_match(MatchResult {
                home_team: "X".to_string(),
                away_team: "Y".to_string(),
                home_goals: 1,
                away_goals: 0,
                round: 1,
                date,
            })
            .unwrap();

        let standings = league.get_standings();
        let last = standings.last().unwrap();

        assert_eq!(last.team, "Z");
        assert_eq!(last.played, 0);
    }

    #[test]
    fn test_head_to_head_tiebreaker() {
        let mut league = League::new();
        let date = Utc.with_ymd_and_hms(2024, 1, 15, 14, 0, 0).unwrap();

        league
            .add_match(MatchResult {
                home_team: "A".to_string(),
                away_team: "B".to_string(),
                home_goals: 1,
                away_goals: 0,
                round: 1,
                date,
            })
            .unwrap();

        league
            .add_match(MatchResult {
                home_team: "A".to_string(),
                away_team: "C".to_string(),
                home_goals: 2,
                away_goals: 1,
                round: 2,
                date,
            })
            .unwrap();

        league
            .add_match(MatchResult {
                home_team: "B".to_string(),
                away_team: "C".to_string(),
                home_goals: 2,
                away_goals: 1,
                round: 3,
                date,
            })
            .unwrap();

        let standings = league.get_standings();

        assert_eq!(standings[0].team, "A");
        assert_eq!(standings[0].points, 6);
        assert_eq!(standings[1].team, "B");
        assert_eq!(standings[1].points, 3);
        assert_eq!(standings[2].team, "C");
        assert_eq!(standings[2].points, 0);
    }
}
