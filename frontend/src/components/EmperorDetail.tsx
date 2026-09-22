import type { Emperor } from '../types'
import { formatRange, formatYear } from '../types'

interface Props {
  emperor: Emperor | null
  onEdit: (e: Emperor) => void
  onDelete: (e: Emperor) => void
}

function Row({ label, value }: { label: string; value: string }) {
  if (!value) return null
  return (
    <div className="detail-row">
      <span className="label">{label}</span>
      <span className="value">{value}</span>
    </div>
  )
}

/** 右侧帝王详情面板。 */
export default function EmperorDetail({ emperor, onEdit, onDelete }: Props) {
  if (!emperor) {
    return (
      <aside className="detail-panel">
        <div className="empty">
          点击左侧图谱或时间轴中的
          <br />
          任意帝王节点，查看详细信息
        </div>
      </aside>
    )
  }

  const title = emperor.templeName || emperor.posthumousName || emperor.name
  const isEmperor = emperor.isEmperor !== false

  return (
    <aside className="detail-panel">
      <div className="detail-header">
        <div className="temple">{title}</div>
        <div className="name">
          {emperor.name}
          {emperor.dynastyName ? ` · ${emperor.dynastyName}` : ''}
        </div>
        <div className={isEmperor ? 'role-tag emperor' : 'role-tag person'}>
          {isEmperor ? '皇帝' : '宗室 · 未即位'}
        </div>
      </div>

      <Row label="庙号" value={emperor.templeName} />
      <Row label="谥号" value={emperor.posthumousName} />
      <Row label="年号" value={emperor.eraNames} />
      <Row label="姓名" value={emperor.name} />
      {isEmperor ? <Row label="在位" value={formatRange(emperor.reignStart, emperor.reignEnd)} /> : null}
      <Row label="生卒" value={`${formatYear(emperor.birthYear)} — ${formatYear(emperor.deathYear)}`} />
      <Row label="世系" value={emperor.relationNote} />

      {emperor.description ? <div className="detail-desc">{emperor.description}</div> : null}

      <div className="detail-actions">
        <button className="btn" onClick={() => onEdit(emperor)}>
          编辑
        </button>
        <button className="btn danger" onClick={() => onDelete(emperor)}>
          删除
        </button>
      </div>
    </aside>
  )
}
