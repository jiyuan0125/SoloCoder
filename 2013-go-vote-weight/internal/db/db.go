package db

import (
	"database/sql"
	"strings"
	"time"

	"votingsystem/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS owners (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		area INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS votings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic TEXT NOT NULL,
		deadline DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		participation REAL DEFAULT 0,
		yes_votes INTEGER DEFAULT 0,
		total_votes INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS votes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		voting_id INTEGER NOT NULL,
		owner_id INTEGER NOT NULL,
		proxy_for_id INTEGER,
		choice TEXT NOT NULL,
		vote_weight INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (voting_id) REFERENCES votings(id),
		FOREIGN KEY (owner_id) REFERENCES owners(id),
		UNIQUE(voting_id, owner_id)
	);

	CREATE TABLE IF NOT EXISTS proxies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		voting_id INTEGER NOT NULL,
		delegator_id INTEGER NOT NULL,
		trustee_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (voting_id) REFERENCES votings(id),
		FOREIGN KEY (delegator_id) REFERENCES owners(id),
		FOREIGN KEY (trustee_id) REFERENCES owners(id),
		UNIQUE(voting_id, delegator_id)
	);

	CREATE TABLE IF NOT EXISTS voting_audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		voting_id INTEGER NOT NULL,
		action_type TEXT NOT NULL,
		owner_id INTEGER,
		details TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (voting_id) REFERENCES votings(id)
	);

	CREATE INDEX IF NOT EXISTS idx_votes_voting_id ON votes(voting_id);
	CREATE INDEX IF NOT EXISTS idx_proxies_voting_id ON proxies(voting_id);
	CREATE INDEX IF NOT EXISTS idx_proxies_trustee ON proxies(voting_id, trustee_id);
	CREATE INDEX IF NOT EXISTS idx_audit_voting_id ON voting_audit_logs(voting_id);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetDB() *sql.DB {
	return s.db
}

func (s *Store) WithTx(fn func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Store) CreateOwner(name string, area int) (*model.Owner, error) {
	res, err := s.db.Exec(`INSERT INTO owners (name, area) VALUES (?, ?)`, name, area)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &model.Owner{ID: id, Name: name, Area: area}, nil
}

func (s *Store) GetOwner(id int64) (*model.Owner, error) {
	row := s.db.QueryRow(`SELECT id, name, area FROM owners WHERE id = ?`, id)
	var o model.Owner
	if err := row.Scan(&o.ID, &o.Name, &o.Area); err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Store) ListOwners() ([]model.Owner, error) {
	rows, err := s.db.Query(`SELECT id, name, area FROM owners ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []model.Owner
	for rows.Next() {
		var o model.Owner
		if err := rows.Scan(&o.ID, &o.Name, &o.Area); err != nil {
			return nil, err
		}
		owners = append(owners, o)
	}
	return owners, nil
}

func (s *Store) CountOwners() (int64, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) FROM owners`)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) GetTotalVotingPower() (int64, error) {
	var total int64
	row := s.db.QueryRow(`SELECT COALESCE(SUM(area), 0) FROM owners`)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Store) CreateVoting(topic string, deadline time.Time) (*model.Voting, error) {
	now := time.Now()
	res, err := s.db.Exec(`INSERT INTO votings (topic, deadline, status, created_at) VALUES (?, ?, ?, ?)`,
		topic, deadline, model.VotingStatusActive, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &model.Voting{
		ID:        id,
		Topic:     topic,
		Deadline:  deadline,
		Status:    model.VotingStatusActive,
		CreatedAt: now,
	}, nil
}

func (s *Store) GetVoting(id int64) (*model.Voting, error) {
	row := s.db.QueryRow(`SELECT id, topic, deadline, status, participation, yes_votes, total_votes, created_at FROM votings WHERE id = ?`, id)
	var v model.Voting
	var status string
	if err := row.Scan(&v.ID, &v.Topic, &v.Deadline, &status, &v.Participation, &v.YesVotes, &v.TotalVotes, &v.CreatedAt); err != nil {
		return nil, err
	}
	v.Status = model.VotingStatus(status)
	return &v, nil
}

func (s *Store) ListVotings() ([]model.Voting, error) {
	rows, err := s.db.Query(`SELECT id, topic, deadline, status, participation, yes_votes, total_votes, created_at FROM votings ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votings []model.Voting
	for rows.Next() {
		var v model.Voting
		var status string
		if err := rows.Scan(&v.ID, &v.Topic, &v.Deadline, &v.CreatedAt, &status, &v.Participation, &v.YesVotes, &v.TotalVotes); err != nil {
			return nil, err
		}
		v.Status = model.VotingStatus(status)
		votings = append(votings, v)
	}
	return votings, nil
}

func (s *Store) ListExpiredActiveVotings() ([]model.Voting, error) {
	now := time.Now()
	rows, err := s.db.Query(`SELECT id, topic, deadline, status, participation, yes_votes, total_votes, created_at FROM votings WHERE status = ? AND deadline < ?`, model.VotingStatusActive, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votings []model.Voting
	for rows.Next() {
		var v model.Voting
		var status string
		if err := rows.Scan(&v.ID, &v.Topic, &v.Deadline, &status, &v.Participation, &v.YesVotes, &v.TotalVotes, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Status = model.VotingStatus(status)
		votings = append(votings, v)
	}
	return votings, nil
}

func (s *Store) UpdateVotingResult(id int64, status model.VotingStatus, participation float64, yesVotes int64, totalVotes int64) error {
	_, err := s.db.Exec(`UPDATE votings SET status = ?, participation = ?, yes_votes = ?, total_votes = ? WHERE id = ?`,
		status, participation, yesVotes, totalVotes, id)
	return err
}

func (s *Store) CreateProxy(tx *sql.Tx, votingID int64, delegatorID int64, trusteeID int64) (*model.Proxy, error) {
	now := time.Now()
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(`INSERT INTO proxies (voting_id, delegator_id, trustee_id, created_at) VALUES (?, ?, ?, ?)`,
			votingID, delegatorID, trusteeID, now)
	} else {
		res, err = s.db.Exec(`INSERT INTO proxies (voting_id, delegator_id, trustee_id, created_at) VALUES (?, ?, ?, ?)`,
			votingID, delegatorID, trusteeID, now)
	}
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &model.Proxy{
		ID:          id,
		VotingID:    votingID,
		DelegatorID: delegatorID,
		TrusteeID:   trusteeID,
		CreatedAt:  now,
	}, nil
}

func (s *Store) GetProxyByDelegator(votingID int64, delegatorID int64) (*model.Proxy, error) {
	row := s.db.QueryRow(`SELECT id, voting_id, delegator_id, trustee_id, created_at FROM proxies WHERE voting_id = ? AND delegator_id = ?`,
		votingID, delegatorID)
	var p model.Proxy
	if err := row.Scan(&p.ID, &p.VotingID, &p.DelegatorID, &p.TrusteeID, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetProxiesByTrustee(votingID int64, trusteeID int64) ([]model.Proxy, error) {
	rows, err := s.db.Query(`SELECT id, voting_id, delegator_id, trustee_id, created_at FROM proxies WHERE voting_id = ? AND trustee_id = ?`,
		votingID, trusteeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []model.Proxy
	for rows.Next() {
		var p model.Proxy
		if err := rows.Scan(&p.ID, &p.VotingID, &p.DelegatorID, &p.TrusteeID, &p.CreatedAt); err != nil {
			return nil, err
		}
		proxies = append(proxies, p)
	}
	return proxies, nil
}

func (s *Store) ListProxies(votingID int64) ([]model.Proxy, error) {
	rows, err := s.db.Query(`SELECT id, voting_id, delegator_id, trustee_id, created_at FROM proxies WHERE voting_id = ? ORDER BY id`, votingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []model.Proxy
	for rows.Next() {
		var p model.Proxy
		if err := rows.Scan(&p.ID, &p.VotingID, &p.DelegatorID, &p.TrusteeID, &p.CreatedAt); err != nil {
			return nil, err
		}
		proxies = append(proxies, p)
	}
	return proxies, nil
}

func (s *Store) CreateVote(tx *sql.Tx, votingID int64, ownerID int64, proxyForID *int64, choice model.VoteChoice, weight int64) (*model.Vote, error) {
	now := time.Now()
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(`INSERT INTO votes (voting_id, owner_id, proxy_for_id, choice, vote_weight, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			votingID, ownerID, proxyForID, choice, weight, now)
	} else {
		res, err = s.db.Exec(`INSERT INTO votes (voting_id, owner_id, proxy_for_id, choice, vote_weight, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			votingID, ownerID, proxyForID, choice, weight, now)
	}
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &model.Vote{
		ID:         id,
		VotingID:   votingID,
		OwnerID:    ownerID,
		ProxyForID: proxyForID,
		Choice:     choice,
		VoteWeight: weight,
		CreatedAt:  now,
	}, nil
}

func (s *Store) GetVote(votingID int64, ownerID int64) (*model.Vote, error) {
	row := s.db.QueryRow(`SELECT id, voting_id, owner_id, proxy_for_id, choice, vote_weight, created_at FROM votes WHERE voting_id = ? AND owner_id = ?`,
		votingID, ownerID)
	var v model.Vote
	if err := row.Scan(&v.ID, &v.VotingID, &v.OwnerID, &v.ProxyForID, &v.Choice, &v.VoteWeight, &v.CreatedAt); err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *Store) ListVotes(votingID int64) ([]model.Vote, error) {
	rows, err := s.db.Query(`SELECT id, voting_id, owner_id, proxy_for_id, choice, vote_weight, created_at FROM votes WHERE voting_id = ? ORDER BY id`, votingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []model.Vote
	for rows.Next() {
		var v model.Vote
		if err := rows.Scan(&v.ID, &v.VotingID, &v.OwnerID, &v.ProxyForID, &v.Choice, &v.VoteWeight, &v.CreatedAt); err != nil {
			return nil, err
		}
		votes = append(votes, v)
	}
	return votes, nil
}

func (s *Store) CalculateVoteStats(votingID int64) (yesVotes int64, totalVotes int64, err error) {
	row := s.db.QueryRow(`SELECT COALESCE(SUM(vote_weight), 0) FROM votes WHERE voting_id = ? AND choice = ?`, votingID, model.VoteChoiceYes)
	if err := row.Scan(&yesVotes); err != nil {
		return 0, 0, err
	}

	row = s.db.QueryRow(`SELECT COALESCE(SUM(vote_weight), 0) FROM votes WHERE voting_id = ?`, votingID)
	if err := row.Scan(&totalVotes); err != nil {
		return 0, 0, err
	}

	return yesVotes, totalVotes, nil
}

func (s *Store) CreateAuditLog(votingID int64, actionType model.VotingActionType, ownerID *int64, details string) error {
	now := time.Now()
	_, err := s.db.Exec(`INSERT INTO voting_audit_logs (voting_id, action_type, owner_id, details, created_at) VALUES (?, ?, ?, ?, ?)`,
		votingID, actionType, ownerID, details, now)
	return err
}

func (s *Store) ListAuditLogs(votingID int64) ([]model.VotingAuditLog, error) {
	rows, err := s.db.Query(`SELECT id, voting_id, action_type, owner_id, details, created_at FROM voting_audit_logs WHERE voting_id = ? ORDER BY id`, votingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.VotingAuditLog
	for rows.Next() {
		var l model.VotingAuditLog
		if err := rows.Scan(&l.ID, &l.VotingID, &l.ActionType, &l.OwnerID, &l.Details, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Store) IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}

func (s *Store) IsUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return false
}

func IsUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint")
}

func (s *Store) GetDelegatees(votingID int64) (map[int64]int64, error) {
	rows, err := s.db.Query(`SELECT trustee_id, COALESCE(SUM(o.area), 0) FROM proxies p JOIN owners o ON p.delegator_id = o.id WHERE p.voting_id = ? GROUP BY trustee_id`, votingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]int64)
	for rows.Next() {
		var trusteeID int64
		var area int64
		if err := rows.Scan(&trusteeID, &area); err != nil {
			return nil, err
		}
		result[trusteeID] = area
	}
	return result, nil
}

func (s *Store) CheckDelegate(votingID int64, ownerID int64) (bool, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM proxies WHERE voting_id = ? AND delegator_id = ?`, votingID, ownerID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) CheckVote(tx *sql.Tx, votingID int64, ownerID int64) (bool, error) {
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id = ? AND owner_id = ?`, votingID, ownerID)
	} else {
		row = s.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id = ? AND owner_id = ?`, votingID, ownerID)
	}
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
