import type { Dynasty } from '../types'

interface Props {
  dynasties: Dynasty[]
  selectedId: number | null
  onSelect: (id: number) => void
  onAdd: () => void
  onEdit: (d: Dynasty) => void
  onDelete: (d: Dynasty) => void
  formatRange: (start: number | null, end: number | null) => string
}

/** 顶部朝代横向选择条，含新增朝代与对当前朝代的编辑/删除入口。 */
export default function DynastyBar({
  dynasties,
  selectedId,
  onSelect,
  onAdd,
  onEdit,
  onDelete,
  formatRange,
}: Props) {
  return (
    <nav className="dynasty-bar">
      {dynasties.map((d) => {
        const active = d.id === selectedId
        return (
          <div
            key={d.id}
            className={active ? 'dynasty-chip active' : 'dynasty-chip'}
            onClick={() => onSelect(d.id)}
            title={d.description}
          >
            <span className="name">{d.name}</span>
            <span className="years">{formatRange(d.startYear, d.endYear)}</span>
            {active ? (
              <span className="chip-actions">
                <button
                  className="chip-btn"
                  title="编辑朝代"
                  onClick={(e) => {
                    e.stopPropagation()
                    onEdit(d)
                  }}
                >
                  ✎
                </button>
                <button
                  className="chip-btn danger"
                  title="删除朝代"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete(d)
                  }}
                >
                  ✕
                </button>
              </span>
            ) : null}
          </div>
        )
      })}
      <button className="btn-add-dynasty" onClick={onAdd} title="新增朝代">
        ＋ 朝代
      </button>
    </nav>
  )
}
