// api.js — обёртки над fetch

async function apiRequest(url, options = {}) {
  const opts = {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  };
  if (opts.body && typeof opts.body === 'object') {
    opts.body = JSON.stringify(opts.body);
  }
  const res = await fetch(url, opts);
  let data = null;
  try { data = await res.json(); } catch (e) { /* нет тела */ }
  if (!res.ok) {
    const err = new Error((data && data.error) || `HTTP ${res.status}`);
    err.status = res.status;
    throw err;
  }
  return data;
}

const API = {
  register: (body) => apiRequest('/api/register', { method: 'POST', body }),
  login:    (body) => apiRequest('/api/login',    { method: 'POST', body }),
  logout:   ()     => apiRequest('/api/logout',   { method: 'POST' }),
  me:       ()     => apiRequest('/api/me'),

  posts:        (cat)   => apiRequest('/api/posts' + (cat ? `?category=${encodeURIComponent(cat)}` : '')),
  createPost:   (body)  => apiRequest('/api/posts', { method: 'POST', body }),
  postByID:     (id)    => apiRequest('/api/posts/' + id),
  createComment:(body)  => apiRequest('/api/comments', { method: 'POST', body }),
  categories:   ()      => apiRequest('/api/categories'),

  users:    ()             => apiRequest('/api/users'),
  messages: (withID, offset, limit = 10) =>
              apiRequest(`/api/messages?with=${withID}&offset=${offset}&limit=${limit}`),
};
