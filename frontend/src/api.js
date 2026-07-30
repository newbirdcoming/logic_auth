const BASE = '/api/v1'

let tokenData = JSON.parse(localStorage.getItem('token') || 'null')

export function getToken() { return tokenData }

export function setToken(data) {
  tokenData = data
  if (data) localStorage.setItem('token', JSON.stringify(data))
  else localStorage.removeItem('token')
}

function headers() {
  const h = { 'Content-Type': 'application/json' }
  if (tokenData) h['Authorization'] = 'Bearer ' + tokenData.access_token
  return h
}

async function tryRefresh() {
  if (!tokenData?.refresh_token) return false
  const res = await fetch(BASE + '/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: tokenData.refresh_token }),
  })
  const d = await res.json()
  if (d.code === 0) {
    tokenData.access_token = d.data.access_token
    tokenData.refresh_token = d.data.refresh_token
    setToken(tokenData)
    return true
  }
  setToken(null)
  return false
}

export async function api(method, path, body) {
  const opt = { method, headers: headers() }
  if (body) opt.body = JSON.stringify(body)
  const res = await fetch(BASE + path, opt)
  const d = await res.json()
  if (!res.ok && d.code === 40103) {
    if (await tryRefresh()) {
      const opt2 = { method, headers: headers() }
      if (body) opt2.body = JSON.stringify(body)
      const res2 = await fetch(BASE + path, opt2)
      return res2.json()
    }
  }
  return d
}

export function isLoggedIn() { return !!tokenData }
