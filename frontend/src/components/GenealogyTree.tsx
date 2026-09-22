import { useMemo } from 'react'
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  type Node,
  type Edge,
  type NodeTypes,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import type { Emperor, TreeNode } from '../types'
import { shortRelation } from '../types'
import EmperorNode from './EmperorNode'

// nodeTypes 需在组件外定义，避免每次渲染重建导致 React Flow 警告。
const nodeTypes: NodeTypes = { emperor: EmperorNode }

const NODE_W = 150
const X_GAP = 34
const Y_GAP = 120

interface Props {
  tree: TreeNode[]
  selectedId: number | null
  onSelect: (e: Emperor) => void
}

/** 把世系森林布局为 React Flow 的节点与连线。 */
function buildGraph(tree: TreeNode[], selectedId: number | null) {
  const nodes: Node[] = []
  const edges: Edge[] = []
  let cursorX = 0

  const visit = (node: TreeNode, depth: number): number => {
    const children = node.children ?? []
    let x: number
    if (children.length === 0) {
      x = cursorX
      cursorX += NODE_W + X_GAP
    } else {
      const xs = children.map((c) => visit(c, depth + 1))
      x = (Math.min(...xs) + Math.max(...xs)) / 2
    }

    const id = String(node.id)
    nodes.push({
      id,
      type: 'emperor',
      position: { x, y: depth * Y_GAP },
      data: { emperor: node as Emperor, active: node.id === selectedId },
    })
    return x
  }

  for (const root of tree) {
    visit(root, 0)
    cursorX += X_GAP * 2 // 森林中多棵树之间留出间隔
    // 建立父子连线
  }

  const addEdges = (node: TreeNode) => {
    for (const c of node.children ?? []) {
      const isLineage = c.edgeType === 'lineage'
      edges.push({
        id: `${node.id}-${c.id}`,
        source: String(node.id),
        target: String(c.id),
        type: 'smoothstep',
        // 隔代/旁系世系边：紫色虚线并标注关系（如“曾孙”）
        label: isLineage ? shortRelation(c.relationNote) : undefined,
        labelStyle: { fill: '#cbbbff', fontSize: 11, fontWeight: 600 },
        labelBgStyle: { fill: '#1a1e27', fillOpacity: 0.92 },
        labelBgPadding: [4, 2] as [number, number],
        labelBgBorderRadius: 4,
        style: isLineage
          ? { stroke: '#8a7bd8', strokeWidth: 1.6, strokeDasharray: '6 4' }
          : { stroke: '#c9a227', strokeWidth: 1.6, opacity: 0.7 },
      })
      addEdges(c)
    }
  }
  for (const root of tree) addEdges(root)

  return { nodes, edges }
}

export default function GenealogyTree({ tree, selectedId, onSelect }: Props) {
  const { nodes, edges } = useMemo(() => buildGraph(tree, selectedId), [tree, selectedId])

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <ReactFlow
        // key 仅在切换朝代时变化，触发重新 fitView；选中节点不重置视图
        key={tree[0]?.dynastyId ?? 'empty'}
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.2}
        onNodeClick={(_, node) => onSelect(node.data.emperor as Emperor)}
      >
        <Background variant={BackgroundVariant.Dots} gap={18} size={1} color="#2b3140" />
        <Controls showInteractive={false} />
        <MiniMap
          pannable
          zoomable
          nodeColor="#c9a227"
          maskColor="rgba(15,17,21,0.75)"
          style={{ background: '#12141a' }}
        />
      </ReactFlow>
      <div className="graph-legend">
        <span className="lg-item">
          <i className="lg-line solid" />父子相传
        </span>
        <span className="lg-item">
          <i className="lg-line dashed" />隔代/旁系世系（如曾孙）
        </span>
        <span className="lg-tip">滚轮缩放 · 拖拽平移 · 点击节点查看/编辑</span>
      </div>
    </div>
  )
}
