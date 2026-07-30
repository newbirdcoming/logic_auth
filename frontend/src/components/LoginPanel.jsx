import { useState } from 'react'
import { api, setToken } from '../api'

export default function LoginPanel({ onLogin }) {
  const [reg, setReg] = useState({ username: '', email: '', password: '', nickname: '' })
  const [login, setLogin] = useState({ username: '', password: '' })
  const [msg, setMsg] = useState('')

  const handleRegister = async () => {
    const d = await api('POST', '/auth/register', reg)
    if (d.code === 0) { setToken(d.data); onLogin(); setMsg('✅ 注册成功') }
    else setMsg('❌ ' + d.message)
  }

  const handleLogin = async () => {
    const d = await api('POST', '/auth/login', login)
    if (d.code === 0) { setToken(d.data); onLogin(); setMsg('✅ 登录成功') }
    else setMsg('❌ ' + d.message)
  }

  const handleLogout = async () => {
    const t = JSON.parse(localStorage.getItem('token') || 'null')
    if (t) await api('POST', '/auth/logout', { refresh_token: t.refresh_token })
    setToken(null)
    onLogin()
    setMsg('✅ 已登出')
  }

  return (
    <div>
      <h2>📝 注册 / 登录</h2>
      {msg && <div className={`status ${msg.startsWith('✅') ? 'success' : 'error'}`}>{msg}</div>}

      <h3>注册</h3>
      <div className="form-row">
        <input placeholder="用户名" value={reg.username} onChange={e => setReg({ ...reg, username: e.target.value })} />
        <input placeholder="邮箱" type="email" value={reg.email} onChange={e => setReg({ ...reg, email: e.target.value })} />
        <input placeholder="密码" type="password" value={reg.password} onChange={e => setReg({ ...reg, password: e.target.value })} />
        <input placeholder="昵称(选填)" value={reg.nickname} onChange={e => setReg({ ...reg, nickname: e.target.value })} />
      </div>
      <button className="btn" onClick={handleRegister}>注册</button>

      <h3 style={{ marginTop: 20 }}>登录</h3>
      <div className="form-row">
        <input placeholder="用户名/邮箱" value={login.username} onChange={e => setLogin({ ...login, username: e.target.value })} />
        <input placeholder="密码" type="password" value={login.password} onChange={e => setLogin({ ...login, password: e.target.value })} />
      </div>
      <button className="btn" onClick={handleLogin}>登录</button>
      <button className="btn danger" style={{ marginLeft: 8 }} onClick={handleLogout}>登出</button>
    </div>
  )
}
