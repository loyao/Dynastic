// 与后端 JSON 结构对应的 TypeScript 类型定义。

export interface Dynasty {
  id: number
  name: string
  startYear: number | null
  endYear: number | null
  capital: string
  description: string
  emperorCount: number
}

export interface Emperor {
  id: number
  dynastyId: number
  dynastyName?: string
  name: string
  templeName: string
  posthumousName: string
  eraNames: string
  fatherId: number | null
  lineageId: number | null
  orderIndex: number
  relationNote: string
  reignStart: number | null
  reignEnd: number | null
  birthYear: number | null
  deathYear: number | null
  description: string
}

export interface TreeNode extends Emperor {
  /** 与父节点的连接方式：father=实线父子，lineage=虚线隔代/旁系，空=根节点 */
  edgeType: 'father' | 'lineage' | ''
  children: TreeNode[]
}

export interface DynastyDetail extends Dynasty {
  emperors: Emperor[]
}

/** 把年份格式化为“公元前 X 年 / 公元 X 年”，null 返回占位符。 */
export function formatYear(year: number | null | undefined): string {
  if (year === null || year === undefined) return '—'
  if (year < 0) return `前${Math.abs(year)}年`
  return `${year}年`
}

/** 生成“前221 — 前207”这类年份区间文本。 */
export function formatRange(start: number | null, end: number | null): string {
  return `${formatYear(start)} — ${formatYear(end)}`
}

/** 从关系说明中提取简短标签，用于图谱连线（取“（”“，”等之前的部分）。 */
export function shortRelation(note: string): string {
  if (!note) return '世系'
  const cut = note.split(/[（(，,]/)[0].trim()
  return cut || '世系'
}
