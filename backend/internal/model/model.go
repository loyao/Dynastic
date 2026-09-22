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

// Emperor 表示一位帝王，是世系图谱的核心节点。
// FatherID  为直接父帝（生父本身是皇帝），在图上以实线父子边连接；
// LineageID 为世系上游帝王（生父非皇帝时，指向最近的帝王祖先，如曾祖），以带标签的虚线边连接；
// 两者皆为 nil 表示该朝代始祖或独立成系。
type Emperor struct {
	ID             int64   `json:"id"`
	DynastyID      int64   `json:"dynastyId"`
	DynastyName    string  `json:"dynastyName,omitempty"`
	Name           string  `json:"name"`           // 姓名
	TempleName     string  `json:"templeName"`     // 庙号，如“太祖”
	PosthumousName string  `json:"posthumousName"` // 谥号，如“高皇帝”
	EraNames       string  `json:"eraNames"`       // 年号，如“洪武”
	FatherID       *int64  `json:"fatherId"`
	LineageID      *int64  `json:"lineageId"`
	OrderIndex     int     `json:"orderIndex"`     // 在位先后顺序
	RelationNote   string  `json:"relationNote"`   // 继位/世系关系说明，如“曾孙”“兄终弟及”
	ReignStart     *int    `json:"reignStart"`     // 在位起始年，负数为公元前
	ReignEnd       *int    `json:"reignEnd"`       // 在位结束年
	BirthYear      *int    `json:"birthYear"`
	DeathYear      *int    `json:"deathYear"`
	Description    string  `json:"description"`
}

// TreeNode 是前端家谱树所需的节点结构。
// EdgeType 描述该节点与父节点的连接方式："father"（实线父子）或 "lineage"（虚线隔代/旁系）。
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
