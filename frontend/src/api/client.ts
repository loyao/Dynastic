import type { Dynasty, DynastyDetail, Emperor, TreeNode } from '../types'

// 通过 Vite 代理访问后端时可使用相对路径；也可用 VITE_API_BASE 指定完整地址。
const BASE = import.meta.env.VITE_API_BASE ?? ''

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (res.status === 204) return undefined as T
  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string }
      if (body.error) msg = body.error
    } catch {
      // 忽略解析失败，使用默认错误信息
    }
    throw new Error(msg)
  }
  return (await res.json()) as T
}

// 新增/编辑朝代时的输入结构（id 由后端生成）。
export type DynastyInput = Omit<Dynasty, 'id' | 'emperorCount'> & { id?: number }
// 新增/编辑帝王时的输入结构。
export type EmperorInput = Omit<Emperor, 'id' | 'dynastyName'> & { id?: number }

export const api = {
  // ---- 查询 ----
  listDynasties: () => request<Dynasty[]>('/api/dynasties'),
  getDynasty: (id: number) => request<DynastyDetail>(`/api/dynasties/${id}`),
  getTree: (dynastyId: number) => request<TreeNode[]>(`/api/dynasties/${dynastyId}/tree`),
  listEmperors: (dynastyId?: number) =>
    request<Emperor[]>(dynastyId ? `/api/emperors?dynastyId=${dynastyId}` : '/api/emperors'),
  getEmperor: (id: number) => request<Emperor>(`/api/emperors/${id}`),

  // ---- 朝代增删改 ----
  createDynasty: (input: DynastyInput) =>
    request<Dynasty>('/api/dynasties', { method: 'POST', body: JSON.stringify(input) }),
  updateDynasty: (id: number, input: DynastyInput) =>
    request<Dynasty>(`/api/dynasties/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  deleteDynasty: (id: number) =>
    request<void>(`/api/dynasties/${id}`, { method: 'DELETE' }),

  // ---- 帝王增删改 ----
  createEmperor: (input: EmperorInput) =>
    request<Emperor>('/api/emperors', { method: 'POST', body: JSON.stringify(input) }),
  updateEmperor: (id: number, input: EmperorInput) =>
    request<Emperor>(`/api/emperors/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  deleteEmperor: (id: number) =>
    request<void>(`/api/emperors/${id}`, { method: 'DELETE' }),
}
