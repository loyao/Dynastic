import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { Emperor } from '../types'
import { formatRange } from '../types'

// React Flow 自定义节点的 data 结构。
export interface EmperorNodeData extends Record<string, unknown> {
  emperor: Emperor
  active: boolean
}

/** 家谱树中的帝王节点。 */
export default function EmperorNode({ data }: NodeProps) {
  const { emperor, active } = data as unknown as EmperorNodeData
  return (
    <div className={active ? 'emperor-node selected' : 'emperor-node'}>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <div className="n-temple">{emperor.templeName || emperor.posthumousName || emperor.name}</div>
      <div className="n-name">{emperor.name}</div>
      <div className="n-years">{formatRange(emperor.reignStart, emperor.reignEnd)}</div>
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
