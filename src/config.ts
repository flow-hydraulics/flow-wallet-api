import { HashAlgorithm, SignatureAlgorithm } from "./lib/flow/crypto"

const chainEmulator = "emulator"
const chainTestnet = "testnet"

const contractsEmulator = {
  FlowToken: "0x0ae53cb6e3f42a79",
  FungibleToken: "0xee82856bf20e2aa6",
  FUSD: "0xf8d6e0586b0a20c7",
}

const contractsTestnet = {
  FlowToken: "0x7e60df042a9c0868",
  FungibleToken: "0x9a0766d93b6608b7",
  FUSD: "0xe223d8a629e49c68",
}

function getContracts(chain) {
  chain = chain || chainEmulator

  switch (chain) {
    case chainEmulator:
      return contractsEmulator
    case chainTestnet:
      return contractsTestnet
  }

  throw `Invalid chain: ${chain}`
}

const config = {
  env: process.env.NODE_ENV,
  port: process.env.PORT || 3000,
  accessApiHost: process.env.ACCESS_API_HOST || "http://localhost:8080",
  adminAddress: process.env.ADMIN_ADDRESS || "0xf8d6e0586b0a20c7",
  adminPrivateKey: process.env.ADMIN_PRIVATE_KEY,
  adminSigAlgo: SignatureAlgorithm[(process.env.ADMIN_SIG_ALGO || "ECDSA_P256")],
  adminHashAlgo: (process.env.ADMIN_HASH_ALGO || "SHA3_256") as HashAlgorithm,
  contracts: getContracts(process.env.CHAIN),
} 

export default config;
