package model

// Dynasty 表示一个朝代。
type Dynasty struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	StartYear   *int   `json:"startYear"`   // 负数表示公元前
	EndYear     *int   `json:"endYear"`     // 负数表示公元前
	Capital     string `json:"capital"`
	Description string `json:"description"`
	EmperorCount int   `json:"emperorCount"`
}

// Emperor 表示世系图谱中的一个人物节点（可以是皇帝，也可以是未即位的宗室祖先）。
// IsEmperor 为 false 时表示非皇帝人物（如戾太子、史皇孙），在图上以虚线框区分。
// FatherID  为直接生父（可为皇帝或非皇帝人物），在图上以实线血脉边连接；隔代/旁系关系通过中间非皇帝祖先节点自然展现。
// 皇位传承箭头由前端按 IsEmperor=true 节点的 OrderIndex 顺序派生，不在模型中存储。
type Emperor struct {
	ID             int64   `json:"id"`
	DynastyID      int64   `json:"dynastyId"`
	DynastyName    string  `json:"dynastyName,omitempty"`
	IsEmperor      bool    `json:"isEmperor"`      // true=皇帝，false=宗室/未即位祖先
	Name           string  `json:"name"`           // 姓名
	TempleName     string  `json:"templeName"`     // 庙号，如“太祖”
	PosthumousName string  `json:"posthumousName"` // 谥号，如“高皇帝”
	EraNames       string  `json:"eraNames"`       // 年号，如“洪武”
	FatherID       *int64  `json:"fatherId"`
	OrderIndex     int     `json:"orderIndex"`     // 在位先后顺序（非皇帝为 0）
	RelationNote   string  `json:"relationNote"`   // 继位/世系关系说明，如“兄终弟及”“仁宗养子”
	ReignStart     *int    `json:"reignStart"`     // 在位起始年，负数为公元前
	ReignEnd       *int    `json:"reignEnd"`       // 在位结束年
	BirthYear      *int    `json:"birthYear"`
	DeathYear      *int    `json:"deathYear"`
	Description    string  `json:"description"`
}

// TreeNode 是前端家谱树所需的节点结构。
// EdgeType 描述该节点与父节点的连接方式："father"（实线血脉），根节点为空。
type TreeNode struct {
	Emperor
	EdgeType string      `json:"edgeType"`
	Children []*TreeNode `json:"children"`
}

// DynastyDetail 返回朝代及其全部帝王，供前端渲染。
type DynastyDetail struct {
	Dynasty
	Emperors []Emperor `json:"emperors"`
}
