package store

import (
	"database/sql"
	"errors"

	"approval-flow/pkg/db"
	"approval-flow/pkg/model"
	"approval-flow/pkg/util"
)

var ErrChainNotFound = errors.New("approval chain not found")
var ErrNodeNotFound = errors.New("approval node not found")

func CreateChain(chain *model.ApprovalChain) error {
	if chain.ID == "" {
		chain.ID = util.NewUUID()
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO approval_chains (id, name, description) VALUES (?, ?, ?)`
	_, err = tx.Exec(query, chain.ID, chain.Name, chain.Description)
	if err != nil {
		return err
	}

	for _, node := range chain.Nodes {
		node.ChainID = chain.ID
		if node.ID == "" {
			node.ID = util.NewUUID()
		}

		approverIDs := util.SliceToJSON(node.ApproverIDs)
		query := `INSERT INTO approval_nodes (id, chain_id, level, node_type, approver_ids, condition, is_sign_all) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.Exec(query, node.ID, node.ChainID, node.Level, node.NodeType, approverIDs, node.Condition, util.BoolToInt(node.IsSignAll))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func UpdateChain(chain *model.ApprovalChain) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE approval_chains SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = tx.Exec(query, chain.Name, chain.Description, chain.ID)
	if err != nil {
		return err
	}

	query = `DELETE FROM approval_nodes WHERE chain_id = ?`
	_, err = tx.Exec(query, chain.ID)
	if err != nil {
		return err
	}

	for _, node := range chain.Nodes {
		node.ChainID = chain.ID
		if node.ID == "" {
			node.ID = util.NewUUID()
		}

		approverIDs := util.SliceToJSON(node.ApproverIDs)
		query := `INSERT INTO approval_nodes (id, chain_id, level, node_type, approver_ids, condition, is_sign_all) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.Exec(query, node.ID, node.ChainID, node.Level, node.NodeType, approverIDs, node.Condition, util.BoolToInt(node.IsSignAll))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetChainByID(id string) (*model.ApprovalChain, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM approval_chains WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	chain := &model.ApprovalChain{}
	err := row.Scan(&chain.ID, &chain.Name, &chain.Description, &chain.CreatedAt, &chain.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChainNotFound
		}
		return nil, err
	}

	nodes, err := getNodesByChainID(chain.ID)
	if err != nil {
		return nil, err
	}
	chain.Nodes = nodes

	return chain, nil
}

func getNodesByChainID(chainID string) ([]*model.ApprovalNode, error) {
	query := `SELECT id, chain_id, level, node_type, approver_ids, condition, is_sign_all, created_at FROM approval_nodes WHERE chain_id = ? ORDER BY level ASC`
	rows, err := db.DB.Query(query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := []*model.ApprovalNode{}
	for rows.Next() {
		node := &model.ApprovalNode{}
		var approverIDs string
		var isSignAll int
		err := rows.Scan(&node.ID, &node.ChainID, &node.Level, &node.NodeType, &approverIDs, &node.Condition, &isSignAll, &node.CreatedAt)
		if err != nil {
			return nil, err
		}
		node.ApproverIDs, _ = util.JSONToSlice[string](approverIDs)
		node.IsSignAll = util.IntToBool(isSignAll)
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func ListChains() ([]*model.ApprovalChain, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM approval_chains`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chains := []*model.ApprovalChain{}
	for rows.Next() {
		chain := &model.ApprovalChain{}
		err := rows.Scan(&chain.ID, &chain.Name, &chain.Description, &chain.CreatedAt, &chain.UpdatedAt)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	return chains, nil
}

func DeleteChain(id string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM approval_nodes WHERE chain_id = ?`
	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}

	query = `DELETE FROM approval_chains WHERE id = ?`
	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
