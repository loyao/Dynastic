import type { Emperor } from '../types'
import { formatRange } from '../types'

interface Props {
  emperors: Emperor[]
  selectedId: number | null
  onSelect: (e: Emperor) => void
}

const CARD_ZONE = 96 // 轴上/下方留给卡片的固定高度，保证圆点对齐

/** 按在位顺序排列的横向时间轴，卡片上下交错分布。 */
export default function Timeline({ emperors, selectedId, onSelect }: Props) {
  const sorted = [...emperors]
    .filter((e) => e.isEmperor !== false)
    .sort((a, b) => a.orderIndex - b.orderIndex || a.id - b.id)

  return (
    <div className="timeline-wrap">
      <div className="timeline">
        <div className="timeline-axis" style={{ top: CARD_ZONE + 5 }} />
        <div className="timeline-items">
          {sorted.map((e, i) => {
            const above = i % 2 === 0
            const title = e.templeName || e.posthumousName || e.name
            const card = (
              <div
                className={e.id === selectedId ? 'tl-card selected' : 'tl-card'}
                onClick={() => onSelect(e)}
              >
                <div className="temple">{title}</div>
                <div className="name">{e.name}</div>
                <div className="reign">{formatRange(e.reignStart, e.reignEnd)}</div>
              </div>
            )
            return (
              <div className="tl-item" key={e.id}>
                <div
                  style={{
                    height: CARD_ZONE,
                    display: 'flex',
                    alignItems: 'flex-end',
                    justifyContent: 'center',
                  }}
                >
                  {above ? card : <div style={{ height: 2, background: '#8c5a3c', opacity: 0.4 }} />}
                </div>
                <div className="tl-dot" />
                <div
                  style={{
                    height: CARD_ZONE,
                    display: 'flex',
                    alignItems: 'flex-start',
                    justifyContent: 'center',
                  }}
                >
                  {!above ? card : <div style={{ height: 2, background: '#8c5a3c', opacity: 0.4 }} />}
                </div>
              </div>
            )
          })}
        </div>
        <div className="hint">按在位先后排列 · 点击卡片查看详情</div>
      </div>
    </div>
  )
}
