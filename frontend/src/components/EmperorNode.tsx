import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { Emperor } from '../types'
import { formatRange, formatYear } from '../types'

// React Flow 自定义节点的 data 结构。
export interface EmperorNodeData extends Record<string, unknown> {
  emperor: Emperor
  active: boolean
}

/** 家谱树中的人物节点：皇帝用实线框，宗室·未即位祖先用虚线框区分。 */
export default function EmperorNode({ data }: NodeProps) {
  const { emperor, active } = data as unknown as EmperorNodeData
  const isEmperor = emperor.isEmperor !== false
  const cls = ['node-card', isEmperor ? 'emperor-node' : 'person-node', active ? 'selected' : '']
    .filter(Boolean)
    .join(' ')

  const title = emperor.templeName || emperor.posthumousName || emperor.name
  const years = isEmperor
    ? formatRange(emperor.reignStart, emperor.reignEnd)
    : emperor.birthYear || emperor.deathYear
      ? `${formatYear(emperor.birthYear)} — ${formatYear(emperor.deathYear)}`
      : '宗室 · 未即位'

  return (
    <div className={cls}>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      {!isEmperor ? <span className="node-badge">未即位</span> : null}
      <div className="n-temple">
        {isEmperor && emperor.orderIndex > 0 ? (
          <span className="n-order" title={`第 ${emperor.orderIndex} 位`}>
            {emperor.orderIndex}
          </span>
        ) : null}
        {title}
      </div>
      <div className="n-name">{emperor.name}</div>
      <div className="n-years">{years}</div>
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
