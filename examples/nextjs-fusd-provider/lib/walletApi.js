const jobPollInterval = 1000
const jobPollAttempts = 40

const jobStatusComplete = "Complete"
const jobStatusError = "Error"

const contentTypeJson = "application/json"

export default class WalletApiClient {

  constructor(baseUrl) {
    this.baseUrl = baseUrl
  }

  async getAccounts() {
    return this.get("/v1/accounts")
  }

  async createAccount() {
    const result = await this.post("/v1/accounts")

    const { jobId } = result;

    const address = await this.pollJobUntilComplete(jobId)

    return address
  }

  async initFungibleToken(address, token) {    
    return this.post(
      `/v1/accounts/${address}/fungible-tokens/${token}`
    )
  }

  async getFungibleToken(address, token) {    
    return this.get(
      `/v1/accounts/${address}/fungible-tokens/${token}`
    )
  }

  async createFungibleTokenWithdrawal(from, to, token, amount) {    
    const result = await this.post(
      `/v1/accounts/${from}/fungible-tokens/${token}/withdrawals`,
      {
        amount,
        recipient: to,
      },
    )

    const { jobId } = result;

    const txId = await this.pollJobUntilComplete(jobId)

    const withdrawal = await this.getFungibleTokenWithdrawal(from, token, txId)

    return withdrawal
  }

  async getFungibleTokenWithdrawal(from, token, id) {    
    return this.get(
      `/v1/accounts/${from}/fungible-tokens/${token}/withdrawals/${id}`
    )
  }

  async getFungibleTokenDeposits(from, token) {    
    return this.get(
      `/v1/accounts/${from}/fungible-tokens/${token}/deposits}`
    )
  }

  async getJob(id) {    
    return this.get(`/v1/jobs/${id}`)
  }

  async pollJobUntilComplete(id) {
    for (let i = 0; i < jobPollAttempts; i++) {
      const job = await this.getJob(id)

      if (job.status === jobStatusError) {
        throw "failed job"
      }
      
      if (job.status === jobStatusComplete) {
        return job.result
      }

      await sleep(jobPollInterval);
    }

    return null
  }

  async post(endpoint, body) {
    return fetch(this.baseUrl + endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': contentTypeJson,
      },
      body: JSON.stringify(body),
    }).then(res => res.json())
  }
  
  async get(endpoint) {
    return fetch(this.baseUrl + endpoint, {
      method: 'GET',
      headers: {
        'Content-Type': contentTypeJson,
      },
    }).then(res => res.json())
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
