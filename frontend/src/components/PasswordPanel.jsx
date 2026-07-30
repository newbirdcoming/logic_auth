import { useState } from 'react'
import { api, setToken, getToken } from '../api'

export default function PasswordPanel() {
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [msg, setMsg] = useState('')

  const handleChange = async () => {
    const d = await api('PUT', '/auth/password', { old_password: oldPwd, new_password: newPwd })
    if (d.code === 0) {
      const t = getToken()
      t.access_token = d.data.access_token
      t.refresh_token = d.data.refresh_token
      setToken(t)
      setMsg('✅ 密码已修改')
    } else {
      setMsg('❌ ' + d.message)
    }
  }

  return (
    <div>
      <h2>🔑 修改密码</h2>
      {msg && <div className={`status ${msg.startsWith('✅') ? 'success' : 'error'}`}>{msg}</div>}
      <div className="form-row">
        <input placeholder="旧密码" type="password" value={oldPwd} onChange={e => setOldPwd(e.target.value)} />
        <input placeholder="新密码" type="password" value={newPwd} onChange={e => setNewPwd(e.target.value)} />
      </div>
      <button className="btn" onClick={handleChange}>修改</button>
    </div>
  )
}
