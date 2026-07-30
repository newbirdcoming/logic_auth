import { useState } from 'react'
import LoginPanel from './components/LoginPanel'
import ProfilePanel from './components/ProfilePanel'
import DevicePanel from './components/DevicePanel'
import PasswordPanel from './components/PasswordPanel'
import AdminPanel from './components/AdminPanel'
import { getToken } from './api'

const TABS = [
  { key: 'login', label: '登录', comp: LoginPanel },
  { key: 'profile', label: '个人信息', comp: ProfilePanel },
  { key: 'devices', label: '设备管理', comp: DevicePanel },
  { key: 'password', label: '修改密码', comp: PasswordPanel },
  { key: 'admin', label: '后台管理', comp: AdminPanel },
]

export default function App() {
  const [tab, setTab] = useState('login')
  const [refresh, setRefresh] = useState(0)
  const token = getToken()

  const handleLogin = () => setRefresh(n => n + 1)

  const ActiveComp = TABS.find(t => t.key === tab)?.comp || LoginPanel

  return (
    <div className="container">
      <h1>🔐 Login Auth Service</h1>
      {token && (
        <div className="token-info">
          <strong>✅ 已登录</strong><br />
          👋 {token.user?.nickname || token.user?.username}
          <br /><small>AT: {token.access_token?.substring(0, 30)}...</small>
        </div>
      )}

      <div className="tabs">
        {TABS.map(t => (
          <div key={t.key} className={`tab ${tab === t.key ? 'active' : ''}`}
            onClick={() => setTab(t.key)}>{t.label}</div>
        ))}
      </div>

      <div className="panel" style={{ display: 'block' }}>
        <ActiveComp key={refresh} onLogin={handleLogin} />
      </div>
    </div>
  )
}
