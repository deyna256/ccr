const CACHE_NAME = 'ccr-v1'
const STATIC_ASSETS = [
  '/',
  '/index.html',
]

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS)
    })
  )
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      )
    })
  )
  self.clients.claim()
})

self.addEventListener('fetch', (event) => {
  const { request } = event
  const url = new URL(request.url)

  if (url.pathname.startsWith('/api/')) {
    event.respondWith(handleAPIRequest(request))
    return
  }

  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request).catch(() => {
        return caches.match('/index.html')
      })
    )
    return
  }

  event.respondWith(
    fetch(request).catch(() => {
      return caches.match(request)
    })
  )
})

async function handleAPIRequest(request) {
  if (request.method === 'GET') {
    try {
      const response = await fetch(request)
      const cache = await caches.open(CACHE_NAME)
      cache.put(request, response.clone())
      return response
    } catch {
      const cached = await caches.match(request)
      if (cached) return cached
      return new Response(JSON.stringify({ error: 'offline' }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' },
      })
    }
  }

  if (['POST', 'PUT', 'DELETE', 'PATCH'].includes(request.method)) {
    try {
      return await fetch(request)
    } catch {
      return new Response(JSON.stringify({ error: 'offline', queued: true }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' },
      })
    }
  }

  return fetch(request)
}

self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-tasks') {
    event.waitUntil(syncTasks())
  }
})

async function syncTasks() {
  try {
    const response = await fetch('/api/sync', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tag: 'sync-tasks' }),
    })
    if (response.ok) {
      const clients = await self.clients.matchAll()
      clients.forEach((client) => {
        client.postMessage({ type: 'SYNC_COMPLETE' })
      })
    }
  } catch {
    // sync failed, will retry on next sync event
  }
}
