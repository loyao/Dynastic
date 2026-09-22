import { useEffect, useState } from 'react'
import Modal from './Modal'
import { api } from '../api/client'
import type { Dynasty, Emperor } from '../types'
import type { EmperorInput } from '../api/client'

interface Props {
  initial: Emperor | null
  defaultDynastyId: number
  dynasties: Dynasty[]
  onClose: () => void
  onSubmit: (input: EmperorInput) => Promise<void>
}

function toNullableInt(s: string): number | null {
  const t = s.trim()
  if (t === '') return null
  const n = Number(t)
  return Number.isNaN(n) ? null : Math.trunc(n)
}

function toStr(v: number | null | undefined): string {
  return v === null || v === undefined ? '' : String(v)
}

/** 人物新增/编辑表单。可在同一弹框内选择所属朝代、身份与父系（直接血缘）。 */
export default function EmperorForm({ initial, defaultDynastyId, dynasties, onClose, onSubmit }: Props) {
  const [dynastyId, setDynastyId] = useState<number>(initial?.dynastyId ?? defaultDynastyId)
  const [isEmperor, setIsEmperor] = useState<boolean>(initial?.isEmperor ?? true)
  const [name, setName] = useState(initial?.name ?? '')
  const [templeName, setTempleName] = useState(initial?.templeName ?? '')
  const [posthumousName, setPosthumousName] = useState(initial?.posthumousName ?? '')
  const [eraNames, setEraNames] = useState(initial?.eraNames ?? '')
  const [orderIndex, setOrderIndex] = useState(toStr(initial?.orderIndex))
  const [fatherId, setFatherId] = useState(toStr(initial?.fatherId))
  const [relationNote, setRelationNote] = useState(initial?.relationNote ?? '')
  const [reignStart, setReignStart] = useState(toStr(initial?.reignStart))
  const [reignEnd, setReignEnd] = useState(toStr(initial?.reignEnd))
  const [birthYear, setBirthYear] = useState(toStr(initial?.birthYear))
  const [deathYear, setDeathYear] = useState(toStr(initial?.deathYear))
  const [description, setDescription] = useState(initial?.description ?? '')

  const [candidates, setCandidates] = useState<Emperor[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isEdit = initial !== null

  // 朝代变化时，加载该朝代全部人物作为「父系」候选（含皇帝与非皇帝祖先）。
  useEffect(() => {
    let alive = true
    api
      .listEmperors(dynastyId)
      .then((list) => {
        if (alive) setCandidates(list)
      })
      .catch(() => {
        if (alive) setCandidates([])
      })
    return () => {
      alive = false
    }
  }, [dynastyId])

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError('帝王姓名不能为空')
      return
    }
    const fid = toNullableInt(fatherId)
    if (isEdit && fid === initial.id) {
      setError('父系不能指向自己')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await onSubmit({
        id: initial?.id,
        dynastyId,
        isEmperor,
        name: name.trim(),
        templeName: templeName.trim(),
        posthumousName: posthumousName.trim(),
        eraNames: eraNames.trim(),
        fatherId: fid,
        orderIndex: isEmperor ? toNullableInt(orderIndex) ?? 0 : 0,
        relationNote: relationNote.trim(),
        reignStart: isEmperor ? toNullableInt(reignStart) : null,
        reignEnd: isEmperor ? toNullableInt(reignEnd) : null,
        birthYear: toNullableInt(birthYear),
        deathYear: toNullableInt(deathYear),
        description: description.trim(),
      })
    } catch (e) {
      setError((e as Error).message)
      setSaving(false)
    }
  }

  // 候选帝王下拉项（编辑时排除自己）
  const options = candidates.filter((c) => !isEdit || c.id !== initial.id)
  const optionLabel = (e: Emperor) =>
    `${e.isEmperor === false ? '〔宗室〕' : `${e.orderIndex}. `}${
      e.templeName || e.posthumousName || e.name
    }（${e.name}）`

  return (
    <Modal
      title={isEdit ? `编辑人物 · ${initial.name}` : '新增人物'}
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
      <div className="field">
        <label>所属朝代 *</label>
        <select value={dynastyId} onChange={(e) => setDynastyId(Number(e.target.value))}>
          {dynasties.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </select>
      </div>
      <div className="field">
        <label>身份 *</label>
        <div className="seg">
          <button
            type="button"
            className={isEmperor ? 'seg-btn active' : 'seg-btn'}
            onClick={() => setIsEmperor(true)}
          >
            皇帝
          </button>
          <button
            type="button"
            className={!isEmperor ? 'seg-btn active' : 'seg-btn'}
            onClick={() => setIsEmperor(false)}
          >
            宗室 · 未即位
          </button>
        </div>
      </div>
      {isEmperor ? (
        <div className="field">
          <label>在位顺序</label>
          <input value={orderIndex} onChange={(e) => setOrderIndex(e.target.value)} placeholder="如 7" />
        </div>
      ) : null}

      <div className="field">
        <label>姓名 *</label>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="如：刘询" />
      </div>
      <div className="field">
        <label>庙号</label>
        <input value={templeName} onChange={(e) => setTempleName(e.target.value)} placeholder="如：中宗" />
      </div>
      <div className="field">
        <label>谥号</label>
        <input
          value={posthumousName}
          onChange={(e) => setPosthumousName(e.target.value)}
          placeholder="如：孝宣皇帝"
        />
      </div>
      <div className="field">
        <label>年号</label>
        <input value={eraNames} onChange={(e) => setEraNames(e.target.value)} placeholder="如：本始" />
      </div>

      <div className="field full">
        <label>父帝 / 父系（直接血缘）</label>
        <select value={fatherId} onChange={(e) => setFatherId(e.target.value)}>
          <option value="">— 无 / 不详（始祖或独立成系） —</option>
          {options.map((e) => (
            <option key={e.id} value={e.id}>
              {optionLabel(e)}
            </option>
          ))}
        </select>
        <span className="hint-text">
          可指向皇帝或非皇帝祖先（如戾太子、史皇孙、赵允让）；隔代/旁系关系会因中间祖先节点而在图上自然展现
        </span>
      </div>

      <div className="field full">
        <label>世系/继位关系说明</label>
        <input
          value={relationNote}
          onChange={(e) => setRelationNote(e.target.value)}
          placeholder="如：武帝曾孙（戾太子之孙）"
        />
      </div>

      <div className="field">
        <label>在位起（年）</label>
        <input
          value={reignStart}
          onChange={(e) => setReignStart(e.target.value)}
          placeholder="公元前填负数"
          disabled={!isEmperor}
        />
      </div>
      <div className="field">
        <label>在位止（年）</label>
        <input value={reignEnd} onChange={(e) => setReignEnd(e.target.value)} disabled={!isEmperor} />
      </div>
      <div className="field">
        <label>生年</label>
        <input value={birthYear} onChange={(e) => setBirthYear(e.target.value)} />
      </div>
      <div className="field">
        <label>卒年</label>
        <input value={deathYear} onChange={(e) => setDeathYear(e.target.value)} />
      </div>

      <div className="field full">
        <label>简介</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>

      {error ? <div className="modal-error">{error}</div> : null}
    </Modal>
  )
}
