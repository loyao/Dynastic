import { useState } from 'react'
import Modal from './Modal'
import type { Dynasty } from '../types'
import type { DynastyInput } from '../api/client'

interface Props {
  initial: Dynasty | null
  onClose: () => void
  onSubmit: (input: DynastyInput) => Promise<void>
}

/** 把输入框字符串解析为可空整数（支持公元前负数），空串返回 null。 */
function toNullableInt(s: string): number | null {
  const t = s.trim()
  if (t === '') return null
  const n = Number(t)
  return Number.isNaN(n) ? null : Math.trunc(n)
}

function toStr(v: number | null | undefined): string {
  return v === null || v === undefined ? '' : String(v)
}

export default function DynastyForm({ initial, onClose, onSubmit }: Props) {
  const [name, setName] = useState(initial?.name ?? '')
  const [startYear, setStartYear] = useState(toStr(initial?.startYear))
  const [endYear, setEndYear] = useState(toStr(initial?.endYear))
  const [capital, setCapital] = useState(initial?.capital ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isEdit = initial !== null

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError('朝代名称不能为空')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await onSubmit({
        id: initial?.id,
        name: name.trim(),
        startYear: toNullableInt(startYear),
        endYear: toNullableInt(endYear),
        capital: capital.trim(),
        description: description.trim(),
      })
    } catch (e) {
      setError((e as Error).message)
      setSaving(false)
    }
  }

  return (
    <Modal
      title={isEdit ? `编辑朝代 · ${initial.name}` : '新增朝代'}
      onClose={onClose}
      footer={
        <>
          <button className="btn ghost" onClick={onClose} disabled={saving}>
            取消
          </button>
          <button className="btn primary" onClick={handleSubmit} disabled={saving}>
            {saving ? '保存中…' : '保存'}
          </button>
        </>
      }
    >
      <div className="field full">
        <label>朝代名称 *</label>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="如：隋朝" />
      </div>
      <div className="field">
        <label>起始年</label>
        <input
          value={startYear}
          onChange={(e) => setStartYear(e.target.value)}
          placeholder="公元前填负数，如 -221"
        />
        <span className="hint-text">公元前用负数表示</span>
      </div>
      <div className="field">
        <label>结束年</label>
        <input value={endYear} onChange={(e) => setEndYear(e.target.value)} placeholder="如 907" />
      </div>
      <div className="field full">
        <label>都城</label>
        <input value={capital} onChange={(e) => setCapital(e.target.value)} placeholder="如：大兴 / 洛阳" />
      </div>
      <div className="field full">
        <label>简介</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="朝代简要说明"
        />
      </div>
      {error ? <div className="modal-error">{error}</div> : null}
    </Modal>
  )
}
