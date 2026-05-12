from datetime import date, time
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Tournament, Team, Match, TeamStatistics, Booking, User
from app.schemas import (
    TournamentCreate, TournamentResponse, TeamCreate, TeamResponse,
    MatchCreate, MatchResponse, MatchUpdateScore, TeamStatisticsResponse
)
from app.services.tournament_service import (
    create_tournament, add_team, generate_schedule,
    book_match_venue, update_match_score, calculate_statistics, forfeit_team
)

router = APIRouter(prefix="/api/tournaments", tags=["赛事管理"])


@router.post("/", response_model=TournamentResponse)
def create_tournament_endpoint(data: TournamentCreate, db: Session = Depends(get_db)):
    try:
        return create_tournament(
            db, data.name, data.venue_type_id,
            data.start_date, data.end_date, data.description
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[TournamentResponse])
def list_tournaments(status: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(Tournament)
    if status:
        query = query.filter(Tournament.status == status)
    return query.order_by(Tournament.created_at.desc()).all()


@router.get("/{tournament_id}", response_model=TournamentResponse)
def get_tournament(tournament_id: int, db: Session = Depends(get_db)):
    tournament = db.query(Tournament).filter(Tournament.id == tournament_id).first()
    if not tournament:
        raise HTTPException(status_code=404, detail="赛事不存在")
    return tournament


@router.post("/{tournament_id}/teams", response_model=TeamResponse)
def add_team_endpoint(tournament_id: int, data: TeamCreate, db: Session = Depends(get_db)):
    try:
        return add_team(
            db, tournament_id, data.name,
            data.captain_name, data.phone
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{tournament_id}/teams", response_model=List[TeamResponse])
def list_teams(tournament_id: int, db: Session = Depends(get_db)):
    return db.query(Team).filter(Team.tournament_id == tournament_id).all()


@router.post("/{tournament_id}/generate-schedule", response_model=List[MatchResponse])
def generate_schedule_endpoint(
    tournament_id: int,
    start_time: time = time(9, 0),
    match_duration_hours: float = 2.0,
    rest_days: int = 1,
    db: Session = Depends(get_db),
):
    try:
        return generate_schedule(
            db, tournament_id, start_time,
            match_duration_hours, rest_days
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{tournament_id}/matches", response_model=List[MatchResponse])
def list_matches(tournament_id: int, round_no: Optional[int] = None,
                  team_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(Match).filter(Match.tournament_id == tournament_id)
    if round_no:
        query = query.filter(Match.round_no == round_no)
    if team_id:
        from sqlalchemy import or_
        query = query.filter(
            or_(Match.home_team_id == team_id, Match.away_team_id == team_id)
        )
    return query.order_by(Match.match_date, Match.start_time).all()


@router.post("/matches/{match_id}/book-venue", response_model=dict)
def book_match_venue_endpoint(match_id: int, venue_id: int, admin_user_id: int,
                               db: Session = Depends(get_db)):
    try:
        booking = book_match_venue(db, match_id, venue_id, admin_user_id)
        return {
            "match_id": match_id,
            "booking_id": booking.id,
            "booking_no": booking.booking_no,
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/matches/{match_id}/score", response_model=MatchResponse)
def update_match_score_endpoint(match_id: int, data: MatchUpdateScore,
                                 db: Session = Depends(get_db)):
    try:
        return update_match_score(db, match_id, data.home_score, data.away_score)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/matches/{match_id}/forfeit")
def forfeit_match_endpoint(match_id: int, forfeit_team_id: int,
                            db: Session = Depends(get_db)):
    try:
        match = update_match_score(
            db, match_id, 0, 0,
            is_forfeit=True, forfeit_team_id=forfeit_team_id
        )
        return {
            "match_id": match.id,
            "status": match.status.value,
            "winner_id": match.winner_id,
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/{tournament_id}/statistics")
def generate_statistics(tournament_id: int, db: Session = Depends(get_db)):
    try:
        stats = calculate_statistics(db, tournament_id)
        result = []
        for s in stats:
            result.append({
                "id": s.id,
                "team_id": s.team_id,
                "team_name": s.team.name if s.team else None,
                "matches_played": s.matches_played,
                "wins": s.wins,
                "losses": s.losses,
                "draws": s.draws,
                "goals_for": s.goals_for,
                "goals_against": s.goals_against,
                "goal_difference": s.goal_difference,
                "points": s.points,
                "rank": s.rank,
            })
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/teams/{team_id}/forfeit")
def forfeit_team_endpoint(team_id: int, is_withdrawn: bool = False,
                           db: Session = Depends(get_db)):
    try:
        team = forfeit_team(db, team_id, is_withdrawn)
        return {
            "team_id": team.id,
            "is_forfeited": team.is_forfeited,
            "is_withdrawn": team.is_withdrawn,
        }
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
