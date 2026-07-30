import { useState } from 'react'
import { api } from '../api'

export default function DevicePanel() {
  const [data, setData] = useState(null)
  const [msg, setMsg] = useState('')

  const fetchDevices = async () => {
    const d = await api('GET', '/users/me/devices')
    setData(d)
  }

  const handleLogoutAll = async () => {
    const d = await api('POST', '/auth/logout/all')
    if (d.code === 0) { setMsg('✅ 已登出所有设备'); fetchDevices() }
  }

  return (
    <div>
      <h2>📱 我的设备</h2>
      {msg && <div className={`status ${msg.startsWith('✅') ? 'success' : 'error'}`}>{msg}</div>}
      <button className="btn" onClick={fetchDevices}>刷新</button>
      <button className="btn danger" style={{ marginLeft: 8 }} onClick={handleLogoutAll}>全部登出</button>
      <pre>{JSON.stringify(data, null, 2)}</pre>
    </div>
  )
}
