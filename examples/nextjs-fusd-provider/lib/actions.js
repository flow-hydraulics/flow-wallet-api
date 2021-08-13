export async function purchaseFusd(recipient, amount) {
  return post('/api/purchase', { recipient, amount })
}

export async function createAccount() {
  return post('/api/accounts')
}

export async function listAccounts() {
  return get('/api/accounts')
}

const contentType = "application/json"

async function get(url) {
  return fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': contentType
    }
  }).then(res => res.json())
}

async function post(url, body) {
  return fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': contentType
    },
    body: JSON.stringify(body),
  }).then(res => res.json())
}
