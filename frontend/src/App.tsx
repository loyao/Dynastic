import { useCallback, useEffect, useState } from 'react'
import { api, type DynastyInput, type EmperorInput } from './api/client'
import type { Dynasty, DynastyDetail, Emperor, TreeNode } from './types'
import { formatRange } from './types'
import DynastyBar from './components/DynastyBar'
import GenealogyTree from './components/GenealogyTree'
import Timeline from './components/Timeline'
import EmperorDetail from './components/EmperorDetail'
import DynastyForm from './components/DynastyForm'
import EmperorForm from './components/EmperorForm'

type ViewMode = 'tree' | 'timeline'

export default function App() {
  const [dynasties, setDynasties] = useState<Dynasty[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [detail, setDetail] = useState<DynastyDetail | null>(null)
  const [tree, setTree] = useState<TreeNode[]>([])
  const [view, setView] = useState<ViewMode>('tree')
  const [selectedEmperor, setSelectedEmperor] = useState<Emperor | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 弹框状态
  const [dynastyModal, setDynastyModal] = useState<{ open: boolean; editing: Dynasty | null }>({
    open: false,
    editing: null,
  })
  const [emperorModal, setEmperorModal] = useState<{ open: boolean; editing: Emperor | null }>({
    open: false,
    editing: null,
  })

  // 拉取当前朝代的详情与世系树。
  const loadCurrent = useCallback((id: number) => {
    return Promise.all([api.getDynasty(id), api.getTree(id)]).then(([d, t]) => {
      setDetail(d)
      setTree(t)
      return d
    })
  }, [])

  // 首次加载：获取朝代列表并默认选中第一个。
  useEffect(() => {
    api
      .listDynasties()
      .then((list) => {
        setDynasties(list)
        if (list.length > 0) setSelectedId(list[0].id)
        else setLoading(false)
      })
      .catch((e: Error) => {
        setError(e.message)
        setLoading(false)
      })
  }, [])

  // 朝代变化：加载详情与世系树。
  useEffect(() => {
    if (selectedId === null) return
    setLoading(true)
    setError(null)
    setSelectedEmperor(null)
    loadCurrent(selectedId)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [selectedId, loadCurrent])

  // ---- 朝代 CRUD ----
  const handleSaveDynasty = async (input: DynastyInput) => {
    if (input.id) {
      await api.updateDynasty(input.id, input)
      const list = await api.listDynasties()
      setDynasties(list)
      await loadCurrent(input.id)
    } else {
      const created = await api.createDynasty(input)
      const list = await api.listDynasties()
      setDynasties(list)
      setDynastyModal({ open: false, editing: null })
      setSelectedId(created.id) // 选中新朝代（会触发 loadCurrent）
      return
    }
    setDynastyModal({ open: false, editing: null })
  }

  const handleDeleteDynasty = async (d: Dynasty) => {
    const count = d.emperorCount ?? 0
    const ok = window.confirm(
      count > 0
        ? `确定删除「${d.name}」及其 ${count} 位帝王吗？此操作不可撤销。`
        : `确定删除「${d.name}」吗？`,
    )
    if (!ok) return
    await api.deleteDynasty(d.id)
    const list = await api.listDynasties()
    setDynasties(list)
    setSelectedEmperor(null)
    setDetail(null)
    setTree([])
    setSelectedId(list.length > 0 ? list[0].id : null)
    if (list.length === 0) setLoading(false)
  }

  // ---- 帝王 CRUD ----
  const handleSaveEmperor = async (input: EmperorInput) => {
    if (input.id) {
      await api.updateEmperor(input.id, input)
    } else {
      await api.createEmperor(input)
    }
    // 若帝王被移动到其他朝代，切换到该朝代再刷新
    if (input.dynastyId !== selectedId) {
      setSelectedId(input.dynastyId)
      setEmperorModal({ open: false, editing: null })
      return
    }
    const d = await loadCurrent(input.dynastyId)
    const list = await api.listDynasties()
    setDynasties(list)
    // 更新右侧详情面板中的当前选中帝王
    if (input.id) {
      const updated = d.emperors.find((e) => e.id === input.id) ?? null
      setSelectedEmperor(updated)
    }
    setEmperorModal({ open: false, editing: null })
  }

  const handleDeleteEmperor = async (e: Emperor) => {
    const ok = window.confirm(`确定删除帝王「${e.name}」吗？其子嗣将退化为独立世系根节点。`)
    if (!ok) return
    await api.deleteEmperor(e.id)
    setSelectedEmperor(null)
    await loadCurrent(e.dynastyId)
    const list = await api.listDynasties()
    setDynasties(list)
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1 className="app-title">
          帝王世系图谱
          <span className="sub">Dynastic Genealogy</span>
        </h1>
        <div className="view-toggle">
          <button className={view === 'tree' ? 'active' : ''} onClick={() => setView('tree')}>
            家谱树
          </button>
          <button className={view === 'timeline' ? 'active' : ''} onClick={() => setView('timeline')}>
            时间轴
          </button>
        </div>
      </header>

      <DynastyBar
        dynasties={dynasties}
        selectedId={selectedId}
        onSelect={setSelectedId}
        onAdd={() => setDynastyModal({ open: true, editing: null })}
        onEdit={(d) => setDynastyModal({ open: true, editing: d })}
        onDelete={handleDeleteDynasty}
        formatRange={formatRange}
      />

      <div className="app-body">
        <div className="view-area">
          {!loading && !error && selectedId !== null ? (
            <div className="view-toolbar">
              <button
                className="btn primary"
                onClick={() => setEmperorModal({ open: true, editing: null })}
              >
                ＋ 新增帝王
              </button>
            </div>
          ) : null}

          {loading ? (
            <div className="loading">加载中…</div>
          ) : error ? (
            <div className="error-box">
              <div>无法连接后端服务</div>
              <div style={{ fontSize: 13, opacity: 0.8 }}>{error}</div>
              <div style={{ fontSize: 13, opacity: 0.7 }}>
                请确认 Go 后端已在 http://localhost:8080 启动，且 MySQL 已连接。
              </div>
            </div>
          ) : view === 'tree' ? (
            <GenealogyTree tree={tree} selectedId={selectedEmperor?.id ?? null} onSelect={setSelectedEmperor} />
          ) : (
            <Timeline
              emperors={detail?.emperors ?? []}
              selectedId={selectedEmperor?.id ?? null}
              onSelect={setSelectedEmperor}
            />
          )}
        </div>

        <EmperorDetail
          emperor={selectedEmperor}
          onEdit={(e) => setEmperorModal({ open: true, editing: e })}
          onDelete={handleDeleteEmperor}
        />
      </div>

      {dynastyModal.open ? (
        <DynastyForm
          initial={dynastyModal.editing}
          onClose={() => setDynastyModal({ open: false, editing: null })}
          onSubmit={handleSaveDynasty}
        />
      ) : null}

      {emperorModal.open && selectedId !== null ? (
        <EmperorForm
          initial={emperorModal.editing}
          defaultDynastyId={emperorModal.editing?.dynastyId ?? selectedId}
          dynasties={dynasties}
          onClose={() => setEmperorModal({ open: false, editing: null })}
          onSubmit={handleSaveEmperor}
        />
      ) : null}
    </div>
  )
}
