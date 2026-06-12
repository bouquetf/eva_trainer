package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type BO struct {
	ID          int       `json:"id"`
	LigueID     int       `json:"ligue_id"`
	Equipe1ID   int       `json:"equipe1_id"`
	Equipe2ID   int       `json:"equipe2_id"`
	VainqueurID *int      `json:"vainqueur_id,omitempty"`
	Format      string    `json:"format"`
	Forfait     int       `json:"forfait"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BOWithDetails struct {
	BO     *BO
	Matchs []*Match `json:"matchs"`
	Bans   []*Ban   `json:"bans"`
	Picks  []*Pick  `json:"picks"`
}

type BOListItem struct {
	BO
	Equipe1Name string `json:"equipe1_name"`
	Equipe2Name string `json:"equipe2_name"`
	DivID       *int   `json:"div_id"`
	DivName     string `json:"div_name"`
}

type boRepo struct {
	db *sql.DB
}

func NewBORepository(db *sql.DB) BORepository {
	return &boRepo{db: db}
}

func scanBO(scanner interface {
	Scan(dest ...interface{}) error
}) (*BO, error) {
	b := &BO{}
	err := scanner.Scan(&b.ID, &b.LigueID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID, &b.Format, &b.Forfait, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func scanBOs(rows *sql.Rows) ([]*BO, error) {
	var bos []*BO
	for rows.Next() {
		b, err := scanBO(rows)
		if err != nil {
			return nil, err
		}
		bos = append(bos, b)
	}
	return bos, rows.Err()
}

func (r *boRepo) Create(ligueID, equipe1ID, equipe2ID int, format string, forfait int) (*BO, error) {
	if format == "" {
		format = "BO3"
	}
	if forfait < 0 || forfait > 2 {
		forfait = 0
	}
	b, err := scanBO(r.db.QueryRow(
		`INSERT INTO bos (ligue_id, equipe1_id, equipe2_id, format, forfait, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, forfait, created_at, updated_at`,
		ligueID, equipe1ID, equipe2ID, format, forfait,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create BO: %w", err)
	}
	return b, nil
}

func (r *boRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM bos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete BO: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBONotFound
	}
	return nil
}

func (r *boRepo) FindByID(id int) (*BO, error) {
	b, err := scanBO(r.db.QueryRow(
		`SELECT id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, forfait, created_at, updated_at
		 FROM bos WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBONotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find BO by id: %w", err)
	}
	return b, nil
}

func (r *boRepo) FindWithDetails(id int) (*BOWithDetails, error) {
	bo, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	matchRepo := &matchRepo{db: r.db}
	matchs, err := matchRepo.FindByBOID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find matchs for BO %d: %w", id, err)
	}

	banRepo := &banRepo{db: r.db}
	bans, err := banRepo.FindByBOID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find bans for BO %d: %w", id, err)
	}

	pickRepo := &pickRepo{db: r.db}
	picks, err := pickRepo.FindByBOID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find picks for BO %d: %w", id, err)
	}

	return &BOWithDetails{
		BO:     bo,
		Matchs: matchs,
		Bans:   bans,
		Picks:  picks,
	}, nil
}

func (r *boRepo) FindByLigueIDWithTeams(ligueID int) ([]*BOListItem, error) {
	rows, err := r.db.Query(
		`SELECT b.id, b.ligue_id, b.equipe1_id, b.equipe2_id, b.vainqueur_id, b.format, b.forfait, b.created_at, b.updated_at,
		        e1.name AS equipe1_name, e2.name AS equipe2_name,
		        d.id AS div_id, COALESCE(d.nom, '') AS div_name
		 FROM bos b
		 JOIN eva_teams e1 ON e1.id = b.equipe1_id
		 JOIN eva_teams e2 ON e2.id = b.equipe2_id
		 LEFT JOIN equipe_eva_div eed ON eed.equipe_eva_id = b.equipe1_id AND eed.ligue_id = b.ligue_id
		 LEFT JOIN divs d ON d.id = eed.div_id
		 WHERE b.ligue_id = $1
		 ORDER BY COALESCE(d.nom, 'ZZZZ') ASC, b.id ASC`,
		ligueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find BOs by ligue with teams: %w", err)
	}
	defer rows.Close()

	var items []*BOListItem
	for rows.Next() {
		item := &BOListItem{}
		var divID sql.NullInt64
		var divName string
		err := rows.Scan(&item.ID, &item.LigueID, &item.Equipe1ID, &item.Equipe2ID, &item.VainqueurID, &item.Format, &item.Forfait, &item.CreatedAt, &item.UpdatedAt,
			&item.Equipe1Name, &item.Equipe2Name, &divID, &divName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan BO list item: %w", err)
		}
		if divID.Valid {
			id := int(divID.Int64)
			item.DivID = &id
		}
		item.DivName = divName
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		return []*BOListItem{}, nil
	}
	return items, nil
}

func (r *boRepo) Update(id int, ligueID, equipe1ID, equipe2ID int, format string, forfait int) (*BO, error) {
	if format == "" {
		format = "BO3"
	}
	if forfait < 0 || forfait > 2 {
		forfait = 0
	}
	b, err := scanBO(r.db.QueryRow(
		`UPDATE bos SET ligue_id = $1, equipe1_id = $2, equipe2_id = $3, format = $4, forfait = $5, updated_at = NOW()
		 WHERE id = $6
		 RETURNING id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, forfait, created_at, updated_at`,
		ligueID, equipe1ID, equipe2ID, format, forfait, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBONotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update BO: %w", err)
	}
	return b, nil
}

func (r *boRepo) RecomputeWinner(id int) (*BO, error) {
	bo, err := r.FindByID(id)
	if err != nil {
		return nil, err
	}

	matchRepo := &matchRepo{db: r.db}
	matchs, err := matchRepo.FindByBOID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find matchs for recompute: %w", err)
	}

	vainqueurID := ComputeBOWinner(matchs, bo.Equipe1ID, bo.Equipe2ID, bo.Forfait)

	b := &BO{}
	err = r.db.QueryRow(
		`UPDATE bos SET vainqueur_id = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, forfait, created_at, updated_at`,
		vainqueurID, id,
	).Scan(&b.ID, &b.LigueID, &b.Equipe1ID, &b.Equipe2ID, &b.VainqueurID, &b.Format, &b.Forfait, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to recompute BO winner: %w", err)
	}
	return b, nil
}

func ComputeBOWinner(matchs []*Match, equipe1ID, equipe2ID int, forfait int) *int {
	if forfait == 1 {
		return &equipe2ID
	}
	if forfait == 2 {
		return &equipe1ID
	}
	wins1, wins2 := 0, 0
	for _, m := range matchs {
		if m.HorsCompetition {
			continue
		}
		if m.VainqueurID != nil {
			switch *m.VainqueurID {
			case equipe1ID:
				wins1++
			case equipe2ID:
				wins2++
			}
		}
	}
	switch {
	case wins1 > wins2:
		return &equipe1ID
	case wins2 > wins1:
		return &equipe2ID
	default:
		return nil
	}
}

func (r *boRepo) FindByLigueID(ligueID int) ([]*BO, error) {
	rows, err := r.db.Query(
		`SELECT id, ligue_id, equipe1_id, equipe2_id, vainqueur_id, format, forfait, created_at, updated_at
		 FROM bos WHERE ligue_id = $1
		 ORDER BY id ASC`,
		ligueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find BOs by ligue: %w", err)
	}
	defer rows.Close()
	return scanBOs(rows)
}
