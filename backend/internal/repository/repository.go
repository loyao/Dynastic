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
	e.era_names, e.father_id, e.lineage_id, e.order_index, e.relation_note,
	e.reign_start, e.reign_end, e.birth_year, e.death_year, e.description`

// scanEmperor 从查询行中读取一位帝王。
func scanEmperor(row interface{ Scan(...any) error }) (model.Emperor, error) {
	var e model.Emperor
	var dynastyName sql.NullString
	err := row.Scan(
		&e.ID, &e.DynastyID, &dynastyName, &e.Name, &e.TempleName, &e.PosthumousName,
		&e.EraNames, &e.FatherID, &e.LineageID, &e.OrderIndex, &e.RelationNote,
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

// BuildTree 依据父子与世系关系把扁平帝王列表组织成世系森林（可能有多个根）。
// 连接优先级：father_id（实线父子）> lineage_id（虚线隔代/旁系）。
// 两者都无法解析到集合内节点时（如始祖），该帝王作为根节点，保证所有节点都出现在树中。
// 内置环检测：若把某节点挂到其后代下会形成环，则退化为根节点，避免死循环。
func BuildTree(emperors []model.Emperor) []*model.TreeNode {
	nodes := make(map[int64]*model.TreeNode, len(emperors))
	for i := range emperors {
		nodes[emperors[i].ID] = &model.TreeNode{Emperor: emperors[i], Children: []*model.TreeNode{}}
	}

	// parentOf 解析一个节点应挂靠的父节点及连接类型。
	parentOf := func(e model.Emperor) (*model.TreeNode, string) {
		if e.FatherID != nil {
			if p, ok := nodes[*e.FatherID]; ok && p.ID != e.ID {
				return p, "father"
			}
		}
		if e.LineageID != nil {
			if p, ok := nodes[*e.LineageID]; ok && p.ID != e.ID {
				return p, "lineage"
			}
		}
		return nil, ""
	}

	// 记录每个节点最终确定的父节点，用于环检测。
	parentLink := make(map[int64]int64, len(emperors))
	// isDescendant 判断 candidate 是否为 node 的后代（沿 parentLink 向上追溯）。
	formsCycle := func(nodeID, candidateParentID int64) bool {
		cur := candidateParentID
		seen := map[int64]bool{}
		for {
			if cur == nodeID {
				return true
			}
			if seen[cur] {
				return false
			}
			seen[cur] = true
			p, ok := parentLink[cur]
			if !ok {
				return false
			}
			cur = p
		}
	}

	roots := make([]*model.TreeNode, 0)
	for i := range emperors {
		e := emperors[i]
		node := nodes[e.ID]
		parent, edgeType := parentOf(e)
		if parent != nil && !formsCycle(e.ID, parent.ID) {
			node.EdgeType = edgeType
			parent.Children = append(parent.Children, node)
			parentLink[e.ID] = parent.ID
			continue
		}
		node.EdgeType = ""
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

// ============ 写操作：朝代 ============

// CreateDynasty 新增朝代，返回新记录 id。sort_order 自动置于末尾。
func (r *Repository) CreateDynasty(d *model.Dynasty) (int64, error) {
	var maxSort sql.NullInt64
	_ = r.db.QueryRow("SELECT MAX(sort_order) FROM dynasty").Scan(&maxSort)
	next := int64(1)
	if maxSort.Valid {
		next = maxSort.Int64 + 1
	}
	res, err := r.db.Exec(`
		INSERT INTO dynasty (name, start_year, end_year, capital, description, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)`,
		d.Name, d.StartYear, d.EndYear, d.Capital, d.Description, next)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateDynasty 更新朝代基本信息。
func (r *Repository) UpdateDynasty(d *model.Dynasty) error {
	_, err := r.db.Exec(`
		UPDATE dynasty SET name=?, start_year=?, end_year=?, capital=?, description=? WHERE id=?`,
		d.Name, d.StartYear, d.EndYear, d.Capital, d.Description, d.ID)
	return err
}

// DeleteDynasty 删除朝代（其下帝王因外键 ON DELETE CASCADE 一并删除）。
func (r *Repository) DeleteDynasty(id int64) error {
	_, err := r.db.Exec("DELETE FROM dynasty WHERE id=?", id)
	return err
}

// ============ 写操作：帝王 ============

// CreateEmperor 新增帝王，返回新记录 id。
func (r *Repository) CreateEmperor(e *model.Emperor) (int64, error) {
	res, err := r.db.Exec(`
		INSERT INTO emperor (dynasty_id, name, temple_name, posthumous_name, era_names,
			father_id, lineage_id, order_index, relation_note,
			reign_start, reign_end, birth_year, death_year, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.DynastyID, e.Name, e.TempleName, e.PosthumousName, e.EraNames,
		e.FatherID, e.LineageID, e.OrderIndex, e.RelationNote,
		e.ReignStart, e.ReignEnd, e.BirthYear, e.DeathYear, e.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateEmperor 更新帝王全部可编辑字段。
func (r *Repository) UpdateEmperor(e *model.Emperor) error {
	_, err := r.db.Exec(`
		UPDATE emperor SET dynasty_id=?, name=?, temple_name=?, posthumous_name=?, era_names=?,
			father_id=?, lineage_id=?, order_index=?, relation_note=?,
			reign_start=?, reign_end=?, birth_year=?, death_year=?, description=?
		WHERE id=?`,
		e.DynastyID, e.Name, e.TempleName, e.PosthumousName, e.EraNames,
		e.FatherID, e.LineageID, e.OrderIndex, e.RelationNote,
		e.ReignStart, e.ReignEnd, e.BirthYear, e.DeathYear, e.Description, e.ID)
	return err
}

// DeleteEmperor 删除帝王，并把指向它的父系/世系引用置空，使其后代退化为根节点而非悬空。
func (r *Repository) DeleteEmperor(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE emperor SET father_id=NULL WHERE father_id=?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE emperor SET lineage_id=NULL WHERE lineage_id=?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM emperor WHERE id=?", id); err != nil {
		return err
	}
	return tx.Commit()
}
