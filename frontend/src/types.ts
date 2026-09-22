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
  /** true=皇帝；false=宗室/未即位祖先（如戾太子、史皇孙） */
  isEmperor: boolean
  name: string
  templeName: string
  posthumousName: string
  eraNames: string
  fatherId: number | null
  orderIndex: number
  relationNote: string
  reignStart: number | null
  reignEnd: number | null
  birthYear: number | null
  deathYear: number | null
  description: string
}

export interface TreeNode extends Emperor {
  /** 与父节点的连接方式：father=实线血脉，空=根节点 */
  edgeType: 'father' | ''
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
