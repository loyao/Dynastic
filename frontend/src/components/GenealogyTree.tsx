import { useMemo, useRef, useState } from 'react'
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  MarkerType,
  getViewportForBounds,
  type Node,
  type Edge,
  type NodeTypes,
  type ReactFlowInstance,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { toPng } from 'html-to-image'
import type { Emperor, TreeNode } from '../types'
import EmperorNode from './EmperorNode'

// nodeTypes 需在组件外定义，避免每次渲染重建导致 React Flow 警告。
const nodeTypes: NodeTypes = { emperor: EmperorNode }

const NODE_W = 150
const NODE_FALLBACK_H = 92 // 导出包围盒在节点未实测时的回退高度
const X_GAP = 34
const Y_GAP = 120

// 宣纸风配色
const COLOR_BLOOD = '#8a6a4f' // 血脉实线：墨褐
const COLOR_SUCCESSION = '#b03a3a' // 传承箭头：暗红
const PAPER_BG = '#f4ecd8' // 导出图片背景（宣纸底）

/** 兄弟（同辈）排序方式：即位顺序 / 出生年 / 默认(id)。仅影响前端布局，不写库。 */
type SortMode = 'order' | 'birth' | 'default'

const SORT_OPTIONS: { value: SortMode; label: string }[] = [
  { value: 'order', label: '即位顺序' },
  { value: 'birth', label: '出生年' },
  { value: 'default', label: '默认(ID)' },
]

// 同辈比较：出生年模式以生年升序（缺生年者排末尾），即位顺序模式以在位序升序，二者均回退到 id。
function compareNodes(a: TreeNode, b: TreeNode, mode: SortMode): number {
  if (mode === 'birth') {
    const ab = a.birthYear ?? Number.POSITIVE_INFINITY
    const bb = b.birthYear ?? Number.POSITIVE_INFINITY
    if (ab !== bb) return ab - bb
    return a.id - b.id
  }
  if (mode === 'order') {
    if (a.orderIndex !== b.orderIndex) return a.orderIndex - b.orderIndex
    return a.id - b.id
  }
  return a.id - b.id
}

// 递归排序世系森林的每一层兄弟，返回新数组（不修改传入的 tree）。
function sortTree(roots: TreeNode[], mode: SortMode): TreeNode[] {
  return roots
    .map((n) => ({ ...n, children: sortTree(n.children ?? [], mode) }))
    .sort((a, b) => compareNodes(a, b, mode))
}

interface Props {
  tree: TreeNode[]
  selectedId: number | null
  onSelect: (e: Emperor) => void
}

/** 把世系森林布局为 React Flow 的节点、血脉（父子）连线与皇位传承箭头。 */
function buildGraph(tree: TreeNode[], selectedId: number | null) {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const all: TreeNode[] = []
  let cursorX = 0

  const visit = (node: TreeNode, depth: number): number => {
    all.push(node)
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
  }

  // 血脉（father 实线）边
  const addEdges = (node: TreeNode) => {
    for (const c of node.children ?? []) {
      edges.push({
        id: `blood-${node.id}-${c.id}`,
        source: String(node.id),
        target: String(c.id),
        type: 'smoothstep',
        style: { stroke: COLOR_BLOOD, strokeWidth: 2, opacity: 0.85 },
      })
      addEdges(c)
    }
  }
  for (const root of tree) addEdges(root)

  // 皇位传承箭头：按在位顺序连接相邻皇帝（跨血脉、跨世代），暗红带箭头。
  const emperors = all
    .filter((n) => n.isEmperor)
    .sort((a, b) => a.orderIndex - b.orderIndex || a.id - b.id)
  for (let i = 1; i < emperors.length; i++) {
    const prev = emperors[i - 1]
    const cur = emperors[i]
    edges.push({
      id: `succ-${prev.id}-${cur.id}`,
      source: String(prev.id),
      target: String(cur.id),
      type: 'default',
      markerEnd: { type: MarkerType.ArrowClosed, color: COLOR_SUCCESSION, width: 18, height: 18 },
      style: { stroke: COLOR_SUCCESSION, strokeWidth: 2.2, opacity: 0.85 },
      zIndex: 1,
    })
  }

  return { nodes, edges }
}

export default function GenealogyTree({ tree, selectedId, onSelect }: Props) {
  const [sortMode, setSortMode] = useState<SortMode>('order')
  const [exporting, setExporting] = useState(false)
  const wrapRef = useRef<HTMLDivElement>(null)
  const rfInstance = useRef<ReactFlowInstance | null>(null)

  const sorted = useMemo(() => sortTree(tree, sortMode), [tree, sortMode])
  const { nodes, edges } = useMemo(() => buildGraph(sorted, selectedId), [sorted, selectedId])

  // 导出整棵世系图为 PNG：手动按全部节点的实测包围盒计算画布尺寸，
  // 并对过宽的世系（如东汉章帝后裔）限制最长边、等比缩小，确保完整导出。
  const handleExport = async () => {
    const wrap = wrapRef.current
    const viewportEl = wrap?.querySelector('.react-flow__viewport') as HTMLElement | null
    if (!viewportEl) return
    // 优先使用 React Flow 运行时节点（含实测宽高）；缺失时回退到布局尺寸，避免被当成点而裁切右/下边缘。
    const liveNodes = rfInstance.current?.getNodes() ?? nodes
    if (liveNodes.length === 0) return

    setExporting(true)
    try {
      let minX = Infinity
      let minY = Infinity
      let maxX = -Infinity
      let maxY = -Infinity
      for (const n of liveNodes) {
        const w = n.measured?.width ?? n.width ?? NODE_W
        const h = n.measured?.height ?? n.height ?? NODE_FALLBACK_H
        const x = n.position.x
        const y = n.position.y
        if (x < minX) minX = x
        if (y < minY) minY = y
        if (x + w > maxX) maxX = x + w
        if (y + h > maxY) maxY = y + h
      }
      const bounds = { x: minX, y: minY, width: maxX - minX, height: maxY - minY }

      const padding = 48
      let imgW = Math.ceil(bounds.width + padding * 2)
      let imgH = Math.ceil(bounds.height + padding * 2)
      // 浏览器 canvas 有最大边长/面积限制，超宽的树按原尺寸导出会被截断，这里限制最长边。
      const MAX_EDGE = 8000
      const k = Math.min(1, MAX_EDGE / Math.max(imgW, imgH))
      imgW = Math.max(1, Math.round(imgW * k))
      imgH = Math.max(1, Math.round(imgH * k))

      const viewport = getViewportForBounds(bounds, imgW, imgH, 0.02, 4, 0.02)
      const dataUrl = await toPng(viewportEl, {
        backgroundColor: PAPER_BG,
        width: imgW,
        height: imgH,
        cacheBust: true,
        style: {
          width: `${imgW}px`,
          height: `${imgH}px`,
          transform: `translate(${viewport.x}px, ${viewport.y}px) scale(${viewport.zoom})`,
        },
      })
      const rawName = tree[0]?.dynastyName || '世系图'
      const safeName = rawName.replace(/[\\/:*?"<>|\s]/g, '')
      const a = document.createElement('a')
      a.download = `${safeName}-世系图.png`
      a.href = dataUrl
      a.click()
    } catch (err) {
      console.error('导出图片失败：', err)
      window.alert('导出失败，请重试。')
    } finally {
      setExporting(false)
    }
  }

  return (
    <div style={{ width: '100%', height: '100%' }} ref={wrapRef}>
      <ReactFlow
        // key 仅在切换朝代时变化，触发重新 fitView；选中节点/排序不重置视图
        key={tree[0]?.dynastyId ?? 'empty'}
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.2}
        onInit={(inst) => {
          rfInstance.current = inst
        }}
        onNodeClick={(_, node) => onSelect(node.data.emperor as Emperor)}
      >
        <Background variant={BackgroundVariant.Dots} gap={20} size={1.4} color="#ddd0b4" />
        <Controls showInteractive={false} />
        <MiniMap
          pannable
          zoomable
          nodeColor={(n) =>
            (n.data as { emperor?: Emperor })?.emperor?.isEmperor === false ? '#b6a68c' : '#8c5a3c'
          }
          maskColor="rgba(244,236,216,0.72)"
          style={{ background: PAPER_BG }}
        />
      </ReactFlow>

      <div className="graph-toolbar">
        <label className="gt-sort">
          <span>兄弟排序</span>
          <select value={sortMode} onChange={(e) => setSortMode(e.target.value as SortMode)}>
            {SORT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
        <button className="btn primary gt-export" onClick={handleExport} disabled={exporting}>
          {exporting ? '导出中…' : '⬇ 导出图片'}
        </button>
      </div>

      <div className="graph-legend">
        <span className="lg-item">
          <i className="lg-line blood" />
          血脉（父子）
        </span>
        <span className="lg-item">
          <i className="lg-line succession" />
          皇位传承
        </span>
        <span className="lg-item">
          <i className="lg-box emperor" />
          皇帝
        </span>
        <span className="lg-item">
          <i className="lg-box person" />
          宗室·未即位
        </span>
        <span className="lg-tip">滚轮缩放 · 拖拽平移 · 点击节点查看/编辑</span>
      </div>
    </div>
  )
}
