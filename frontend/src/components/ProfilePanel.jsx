import { useState } from 'react'
import { api, getToken } from '../api'

export default function ProfilePanel() {
  const [data, setData] = useState(null)
  const [nickname, setNickname] = useState('')
  const [email, setEmail] = useState('')
  const [msg, setMsg] = useState('')

  const fetchProfile = async () => {
    const d = await api('GET', '/users/me')
    setData(d)
  }

  const handleUpdate = async () => {
    const body = {}
    if (nickname) body.nickname = nickname
    if (email) body.email = email
    const d = await api('PUT', '/users/me', body)
    setMsg(d.code === 0 ? '✅ 已更新' : '❌ ' + d.message)
  }

  const user = getToken()?.user

  return (
    <div>
      <h2>👤 个人信息</h2>
      {msg && <div className={`status ${msg.startsWith('✅') ? 'success' : 'error'}`}>{msg}</div>}
      {user && <div className="user-badge">👋 {user.nickname || user.username}</div>}
      <button className="btn" onClick={fetchProfile}>刷新</button>
      <pre>{JSON.stringify(data, null, 2)}</pre>

      <h3 style={{ marginTop: 16 }}>修改资料</h3>
      <div className="form-row">
        <input placeholder="新昵称" value={nickname} onChange={e => setNickname(e.target.value)} />
        <input placeholder="新邮箱" type="email" value={email} onChange={e => setEmail(e.target.value)} />
      </div>
      <button className="btn" onClick={handleUpdate}>保存</button>
    </div>
  )
}
