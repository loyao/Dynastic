package repository

import (
	"database/sql"
	"fmt"
	"sort"

	"dynastic/internal/model"
)

// Repository 封装所有数据库访问。
type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const emperorColumns = `e.id, e.dynasty_id, d.name, e.name, e.temple_name, e.posthumous_name,
	e.era_names, e.father_id, e.order_index, e.relation_note,
	e.reign_start, e.reign_end, e.birth_year, e.death_year, e.description`

// scanEmperor 从查询行中读取一位帝王。
func scanEmperor(row interface{ Scan(...any) error }) (model.Emperor, error) {
	var e model.Emperor
	var dynastyName sql.NullString
	err := row.Scan(
		&e.ID, &e.DynastyID, &dynastyName, &e.Name, &e.TempleName, &e.PosthumousName,
		&e.EraNames, &e.FatherID, &e.OrderIndex, &e.RelationNote,
		&e.ReignStart, &e.ReignEnd, &e.BirthYear, &e.DeathYear, &e.Description,
	)
	if err != nil {
		return e, err
	}
	e.DynastyName = dynastyName.String
	return e, err
}

// ListDynasties 返回全部朝代（含帝王数量），按 sort_order 升序。
func (r *Repository) ListDynasties() ([]model.Dynasty, error) {
	rows, err := r.db.Query(`
		SELECT d.id, d.name, d.start_year, d.end_year, d.capital, d.description,
		       (SELECT COUNT(*) FROM emperor e WHERE e.dynasty_id = d.id) AS emperor_count
		FROM dynasty d
		ORDER BY d.sort_order ASC, d.id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Dynasty, 0)
	for rows.Next() {
		var d model.Dynasty
		if err := rows.Scan(&d.ID, &d.Name, &d.StartYear, &d.EndYear, &d.Capital, &d.Description, &d.EmperorCount); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetDynasty 返回朝代基本信息。
func (r *Repository) GetDynasty(id int64) (*model.Dynasty, error) {
	var d model.Dynasty
	err := r.db.QueryRow(`
		SELECT d.id, d.name, d.start_year, d.end_year, d.capital, d.description,
		       (SELECT COUNT(*) FROM emperor e WHERE e.dynasty_id = d.id)
		FROM dynasty d WHERE d.id = ?`, id).
		Scan(&d.ID, &d.Name, &d.StartYear, &d.EndYear, &d.Capital, &d.Description, &d.EmperorCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListEmperors 返回帝王列表。dynastyID 为 nil 时返回全部，否则按朝代过滤；按在位顺序排序。
func (r *Repository) ListEmperors(dynastyID *int64) ([]model.Emperor, error) {
	query := fmt.Sprintf(`SELECT %s FROM emperor e JOIN dynasty d ON d.id = e.dynasty_id`, emperorColumns)
	args := []any{}
	if dynastyID != nil {
		query += " WHERE e.dynasty_id = ?"
		args = append(args, *dynastyID)
	}
	query += " ORDER BY e.dynasty_id ASC, e.order_index ASC, e.id ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Emperor, 0)
	for rows.Next() {
		e, err := scanEmperor(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEmperor 返回单个帝王详情。
func (r *Repository) GetEmperor(id int64) (*model.Emperor, error) {
	query := fmt.Sprintf(`SELECT %s FROM emperor e JOIN dynasty d ON d.id = e.dynasty_id WHERE e.id = ?`, emperorColumns)
	row := r.db.QueryRow(query, id)
	e, err := scanEmperor(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetDynastyDetail 返回朝代及其全部帝王。
func (r *Repository) GetDynastyDetail(id int64) (*model.DynastyDetail, error) {
	d, err := r.GetDynasty(id)
	if err != nil || d == nil {
		return nil, err
	}
	emperors, err := r.ListEmperors(&id)
	if err != nil {
		return nil, err
	}
	return &model.DynastyDetail{Dynasty: *d, Emperors: emperors}, nil
}

// BuildTree 依据父子关系把扁平帝王列表组织成世系森林（可能有多个根）。
// 找不到父节点的帝王（如养子继位、隔代继承）会被视为根节点，保证所有节点都出现在树中。
func BuildTree(emperors []model.Emperor) []*model.TreeNode {
	nodes := make(map[int64]*model.TreeNode, len(emperors))
	for i := range emperors {
		nodes[emperors[i].ID] = &model.TreeNode{Emperor: emperors[i], Children: []*model.TreeNode{}}
	}

	roots := make([]*model.TreeNode, 0)
	for i := range emperors {
		e := emperors[i]
		node := nodes[e.ID]
		if e.FatherID != nil {
			if parent, ok := nodes[*e.FatherID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	// 兄弟之间按在位顺序排序，根节点同样排序。
	var sortNodes func(list []*model.TreeNode)
	sortNodes = func(list []*model.TreeNode) {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].OrderIndex != list[j].OrderIndex {
				return list[i].OrderIndex < list[j].OrderIndex
			}
			return list[i].ID < list[j].ID
		})
		for _, n := range list {
			sortNodes(n.Children)
		}
	}
	sortNodes(roots)
	return roots
}

// GetTree 返回指定朝代的世系森林。
func (r *Repository) GetTree(dynastyID int64) ([]*model.TreeNode, error) {
	emperors, err := r.ListEmperors(&dynastyID)
	if err != nil {
		return nil, err
	}
	return BuildTree(emperors), nil
}
