import * as dotenv from "dotenv"
dotenv.config()

import {HashAlgorithm, SignatureAlgorithm} from "src/lib/crypto"
import {KeyType} from "src/lib/keys"

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

const defaultSigAlgo = SignatureAlgorithm.ECDSA_P256
const defaultHashAlgo = HashAlgorithm.SHA3_256
const defaultkeyType = KeyType.InMemory

function parseSigAlgo(sigAlgo: string): SignatureAlgorithm {
  return SignatureAlgorithm[sigAlgo] || defaultSigAlgo
}

function parseHashAlgo(hashAlgo: string): HashAlgorithm {
  return HashAlgorithm[hashAlgo] || defaultHashAlgo
}

function parseKeyType(keyType: string): KeyType {
  return KeyType[keyType] || defaultkeyType
}

function parseHex(hex: string): Buffer | null {
  if (!hex) {
    return null
  }

  return Buffer.from(hex, "hex")
}

const config = {
  env: process.env.NODE_ENV,
  port: process.env.PORT || 3000,
  accessApiHost: process.env.ACCESS_API_HOST || "http://localhost:8080",

  adminAddress: process.env.ADMIN_ADDRESS || "0xf8d6e0586b0a20c7",
  adminKeyType: parseKeyType(process.env.ADMIN_KEY_TYPE),
  adminPrivateKey: process.env.ADMIN_PRIVATE_KEY,
  adminSigAlgo: parseSigAlgo(process.env.ADMIN_SIG_ALGO),
  adminHashAlgo: parseHashAlgo(process.env.ADMIN_HASH_ALGO),

  userKeyType: parseKeyType(process.env.USER_KEY_TYPE),
  userSigAlgo: parseSigAlgo(process.env.USER_SIG_ALGO),
  userHashAlgo: parseHashAlgo(process.env.USER_HASH_ALGO),
  userEncryptionKey: parseHex(process.env.USER_ENCRYPTION_KEY),

  contracts: getContracts(process.env.CHAIN),
}

export default config
