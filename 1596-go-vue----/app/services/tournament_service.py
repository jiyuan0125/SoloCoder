from datetime import datetime, date, time, timedelta
from typing import List, Optional, Dict, Any
from sqlalchemy.orm import Session

from app.models import (
    Tournament, Team, Match, Booking, Venue, VenueType, User, TeamStatistics, Notification,
    TournamentStatus, MatchStatus, BookingStatus, BookingType
)
from app.utils import generate_booking_no, round_robin_schedule
from app.config import settings


def create_tournament(db: Session, name: str, venue_type_id: int,
                      start_date: date, end_date: date,
                      description: Optional[str] = None) -> Tournament:
    if start_date > end_date:
        raise ValueError("开始日期不能晚于结束日期")
    
    tournament = Tournament(
        name=name,
        venue_type_id=venue_type_id,
        start_date=start_date,
        end_date=end_date,
        description=description,
        status=TournamentStatus.DRAFT,
    )
    db.add(tournament)
    db.commit()
    db.refresh(tournament)
    
    return tournament


def add_team(db: Session, tournament_id: int, name: str,
             captain_name: Optional[str] = None,
             phone: Optional[str] = None) -> Team:
    tournament = db.query(Tournament).filter(Tournament.id == tournament_id).first()
    if not tournament:
        raise ValueError("赛事不存在")
    
    if tournament.status != TournamentStatus.DRAFT:
        raise ValueError("赛事已开始，无法添加队伍")
    
    existing = db.query(Team).filter(
        Team.tournament_id == tournament_id,
        Team.name == name,
    ).first()
    if existing:
        raise ValueError("该赛事中已存在同名队伍")
    
    team = Team(
        tournament_id=tournament_id,
        name=name,
        captain_name=captain_name,
        phone=phone,
    )
    db.add(team)
    db.commit()
    db.refresh(team)
    
    return team


def check_team_conflict(db: Session, team_id: int, match_date: date,
                        start_time: time, end_time: time,
                        exclude_match_id: Optional[int] = None) -> bool:
    team = db.query(Team).filter(Team.id == team_id).first()
    if not team:
        return False
    
    team_matches = db.query(Match).filter(
        Match.tournament_id == team.tournament_id,
        Match.match_date == match_date,
        Match.status.in_([MatchStatus.SCHEDULED, MatchStatus.IN_PROGRESS, MatchStatus.COMPLETED]),
        or_(Match.home_team_id == team_id, Match.away_team_id == team_id),
    )
    if exclude_match_id:
        team_matches = team_matches.filter(Match.id != exclude_match_id)
    
    for match in team_matches.all():
        match_start = match.start_time
        match_end = match.end_time
        
        start_dt = datetime.combine(match_date, start_time)
        end_dt = datetime.combine(match_date, end_time)
        m_start_dt = datetime.combine(match_date, match_start)
        m_end_dt = datetime.combine(match_date, match_end)
        
        min_interval = timedelta(hours=settings.MIN_MATCH_INTERVAL_HOURS)
        
        if start_dt < m_end_dt + min_interval and end_dt + min_interval > m_start_dt:
            return True
    
    return False


def generate_schedule(db: Session, tournament_id: int,
                      start_time: time = time(9, 0),
                      match_duration_hours: float = 2.0,
                      rest_days: int = 1) -> List[Match]:
    tournament = db.query(Tournament).filter(Tournament.id == tournament_id).first()
    if not tournament:
        raise ValueError("赛事不存在")
    
    teams = db.query(Team).filter(
        Team.tournament_id == tournament_id,
        Team.is_forfeited == False,
        Team.is_withdrawn == False,
    ).all()
    
    if len(teams) < 2:
        raise ValueError("至少需要2支队伍才能生成赛程")
    
    team_ids = [t.id for t in teams]
    
    matches_schedule = round_robin_schedule(
        teams=team_ids,
        start_date=tournament.start_date,
        start_time=start_time,
        match_duration_hours=match_duration_hours,
        rest_days=rest_days,
    )
    
    created_matches = []
    
    for schedule_item in matches_schedule:
        match_date = schedule_item["match_date"]
        st = schedule_item["start_time"]
        duration = schedule_item["duration_hours"]
        end_dt = datetime.combine(match_date, st) + timedelta(hours=duration)
        et = end_dt.time()
        
        if schedule_item["home_team_id"]:
            if check_team_conflict(db, schedule_item["home_team_id"], match_date, st, et):
                continue
        
        if schedule_item["away_team_id"]:
            if check_team_conflict(db, schedule_item["away_team_id"], match_date, st, et):
                continue
        
        match = Match(
            tournament_id=tournament_id,
            match_no=schedule_item["match_no"],
            round_no=schedule_item["round_no"],
            home_team_id=schedule_item["home_team_id"],
            away_team_id=schedule_item["away_team_id"],
            match_date=match_date,
            start_time=st,
            end_time=et,
            status=MatchStatus.SCHEDULED,
        )
        db.add(match)
        db.flush()
        created_matches.append(match)
    
    tournament.status = TournamentStatus.SCHEDULED
    db.commit()
    
    for match in created_matches:
        db.refresh(match)
    
    return created_matches


def book_match_venue(db: Session, match_id: int, venue_id: int, admin_user_id: int) -> Booking:
    match = db.query(Match).filter(Match.id == match_id).first()
    if not match:
        raise ValueError("比赛不存在")
    
    if match.booking_id:
        raise ValueError("该比赛已有场地预约")
    
    from app.services.booking_service import check_venue_availability
    
    if not check_venue_availability(db, venue_id, match.match_date, match.start_time, match.end_time):
        raise ValueError("该时段场地不可用")
    
    booking_no = generate_booking_no()
    
    booking = Booking(
        booking_no=booking_no,
        user_id=admin_user_id,
        venue_id=venue_id,
        booking_type=BookingType.MATCH,
        booking_date=match.match_date,
        start_time=match.start_time,
        end_time=match.end_time,
        hours=(datetime.combine(match.match_date, match.end_time) - 
               datetime.combine(match.match_date, match.start_time)).total_seconds() / 3600,
        original_amount=0,
        discount_amount=0,
        final_amount=0,
        status=BookingStatus.CONFIRMED,
        paid_at=datetime.now(),
        notes=f"比赛: {match.match_no}",
    )
    
    db.add(booking)
    db.flush()
    
    match.booking_id = booking.id
    db.commit()
    db.refresh(booking)
    db.refresh(match)
    
    return booking


def update_match_score(db: Session, match_id: int, home_score: int, away_score: int,
                        is_forfeit: bool = False, forfeit_team_id: Optional[int] = None) -> Match:
    match = db.query(Match).filter(Match.id == match_id).first()
    if not match:
        raise ValueError("比赛不存在")
    
    if is_forfeit:
        if not forfeit_team_id:
            raise ValueError("弃权比赛需指定弃权队伍")
        
        match.is_forfeit = True
        match.forfeit_team_id = forfeit_team_id
        match.status = MatchStatus.FORFEIT
        
        if forfeit_team_id == match.home_team_id:
            match.winner_id = match.away_team_id
        else:
            match.winner_id = match.home_team_id
    else:
        if home_score < 0 or away_score < 0:
            raise ValueError("比分不能为负数")
        
        match.home_score = home_score
        match.away_score = away_score
        match.status = MatchStatus.COMPLETED
        
        if home_score > away_score:
            match.winner_id = match.home_team_id
        elif away_score > home_score:
            match.winner_id = match.away_team_id
    
    db.commit()
    db.refresh(match)
    
    return match


def calculate_statistics(db: Session, tournament_id: int) -> List[TeamStatistics]:
    tournament = db.query(Tournament).filter(Tournament.id == tournament_id).first()
    if not tournament:
        raise ValueError("赛事不存在")
    
    teams = db.query(Team).filter(Team.tournament_id == tournament_id).all()
    stats_list = []
    
    for team in teams:
        stats = db.query(TeamStatistics).filter(
            TeamStatistics.team_id == team.id,
            TeamStatistics.tournament_id == tournament_id,
        ).first()
        
        if not stats:
            stats = TeamStatistics(
                tournament_id=tournament_id,
                team_id=team.id,
            )
            db.add(stats)
        
        if team.is_forfeited or team.is_withdrawn:
            stats.matches_played = None
            stats.wins = None
            stats.losses = None
            stats.draws = None
            stats.goals_for = None
            stats.goals_against = None
            stats.goal_difference = None
            stats.points = None
            stats.rank = None
            stats_list.append(stats)
            continue
        
        valid_matches = db.query(Match).filter(
            Match.tournament_id == tournament_id,
            Match.status.notin_([MatchStatus.CANCELLED, MatchStatus.FORFEIT]),
            Match.home_score.isnot(None),
            Match.away_score.isnot(None),
            or_(Match.home_team_id == team.id, Match.away_team_id == team.id),
        ).all()
        
        matches_played = 0
        wins = 0
        losses = 0
        draws = 0
        goals_for = 0
        goals_against = 0
        
        for match in valid_matches:
            if match.is_forfeit:
                continue
            
            if match.home_team_id == team.id and match.away_team_id == team.id:
                continue
            
            matches_played += 1
            
            if match.home_team_id == team.id:
                gf = match.home_score or 0
                ga = match.away_score or 0
            else:
                gf = match.away_score or 0
                ga = match.home_score or 0
            
            goals_for += gf
            goals_against += ga
            
            if gf > ga:
                wins += 1
            elif gf < ga:
                losses += 1
            else:
                draws += 1
        
        stats.matches_played = matches_played
        stats.wins = wins
        stats.losses = losses
        stats.draws = draws
        stats.goals_for = goals_for
        stats.goals_against = goals_against
        stats.goal_difference = goals_for - goals_against
        stats.points = wins * 3 + draws
        
        stats_list.append(stats)
    
    db.flush()
    
    active_stats = [s for s in stats_list if s.points is not None]
    active_stats.sort(key=lambda x: (-x.points, -(x.goal_difference or 0), -(x.goals_for or 0)))
    
    for rank, s in enumerate(active_stats, 1):
        s.rank = rank
    
    tournament.status = TournamentStatus.COMPLETED
    db.commit()
    
    return stats_list


def forfeit_team(db: Session, team_id: int, is_withdrawn: bool = False) -> Team:
    team = db.query(Team).filter(Team.id == team_id).first()
    if not team:
        raise ValueError("队伍不存在")
    
    team.is_forfeited = True
    if is_withdrawn:
        team.is_withdrawn = True
    
    db.commit()
    db.refresh(team)
    
    return team
