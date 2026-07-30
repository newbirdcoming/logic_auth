import { useState } from 'react'
import { api } from '../api'

export default function AdminPanel() {
  const [tab, setTab] = useState('users')
  const [users, setUsers] = useState([])
  const [roles, setRoles] = useState([])
  const [perms, setPerms] = useState([])
  const [msg, setMsg] = useState('')

  const fetchUsers = async () => {
    const d = await api('GET', '/admin/users?page=1&page_size=50')
    if (d.code === 0) setUsers(d.data?.list || [])
  }
  const toggleUser = async (id, status) => {
    const d = await api('PUT', `/admin/users/${id}/status`, { status })
    if (d.code === 0) { setMsg('✅ 已更新'); fetchUsers() }
  }
  const fetchRoles = async () => {
    const d = await api('GET', '/admin/roles')
    if (d.code === 0) setRoles(d.data || [])
  }
  const createRole = async () => {
    const name = prompt('角色名称：')
    if (!name) return
    const d = await api('POST', '/admin/roles', { name })
    if (d.code === 0) { setMsg('✅ 已创建'); fetchRoles() }
  }
  const deleteRole = async (id) => {
    if (!confirm('确定删除？')) return
    const d = await api('DELETE', '/admin/roles/' + id)
    if (d.code === 0) { setMsg('✅ 已删除'); fetchRoles() }
  }
  const fetchPerms = async () => {
    const d = await api('GET', '/admin/permissions')
    if (d.code === 0) setPerms(d.data || [])
  }
  const createPerm = async () => {
    const code = prompt('权限编码 (如 user:create)：')
    if (!code) return
    const d = await api('POST', '/admin/permissions', { code, name: code, module: code.split(':')[0] || 'system' })
    if (d.code === 0) { setMsg('✅ 已创建'); fetchPerms() }
  }

  return (
    <div>
      <h2>⚙️ 后台管理</h2>
      {msg && <div className={`status ${msg.startsWith('✅') ? 'success' : 'error'}`}>{msg}</div>}
      <div className="tabs" style={{ marginBottom: 10 }}>
        <div className={`tab ${tab === 'users' ? 'active' : ''}`} onClick={() => { setTab('users'); fetchUsers() }}>用户</div>
        <div className={`tab ${tab === 'roles' ? 'active' : ''}`} onClick={() => { setTab('roles'); fetchRoles() }}>角色</div>
        <div className={`tab ${tab === 'perms' ? 'active' : ''}`} onClick={() => { setTab('perms'); fetchPerms() }}>权限</div>
      </div>

      {tab === 'users' && (
        <div>
          <button className="btn" onClick={fetchUsers}>刷新用户列表</button>
          <table><thead><tr><th>ID</th><th>用户名</th><th>邮箱</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>{users.map(u => (
              <tr key={u.ID}><td>{u.ID}</td><td>{u.Username}</td><td>{u.Email}</td>
                <td>{u.Status === 1 ? '正常' : '禁用'}</td>
                <td><button className={`btn sm ${u.Status === 1 ? 'danger' : ''}`}
                  onClick={() => toggleUser(u.ID, u.Status === 1 ? 0 : 1)}>{u.Status === 1 ? '禁用' : '启用'}</button></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}

      {tab === 'roles' && (
        <div>
          <button className="btn" onClick={createRole}>创建角色</button>
          <button className="btn" style={{ marginLeft: 8 }} onClick={fetchRoles}>刷新</button>
          <table><thead><tr><th>ID</th><th>名称</th><th>描述</th><th>操作</th></tr></thead>
            <tbody>{roles.map(r => (
              <tr key={r.ID}><td>{r.ID}</td><td>{r.Name}</td><td>{r.Description || '-'}</td>
                <td>{r.IsSystem ? '系统内置' : <button className="btn sm danger" onClick={() => deleteRole(r.ID)}>删除</button>}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}

      {tab === 'perms' && (
        <div>
          <button className="btn" onClick={createPerm}>创建权限</button>
          <button className="btn" style={{ marginLeft: 8 }} onClick={fetchPerms}>刷新</button>
          <table><thead><tr><th>ID</th><th>编码</th><th>模块</th></tr></thead>
            <tbody>{perms.map(p => (
              <tr key={p.ID}><td>{p.ID}</td><td>{p.Code}</td><td>{p.Module}</td></tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </div>
  )
}
