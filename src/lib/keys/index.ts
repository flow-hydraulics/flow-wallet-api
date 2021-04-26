import * as Crypto from "../crypto"

export enum KeyType {
  InMemory = "in-memory",
  GoogleKMS = "google-kms",
}

export interface Key {
  getPublicKey(): Crypto.PublicKey
  getSignatureAlgorithm(): Crypto.SignatureAlgorithm
  getHashAlgorithm(): Crypto.HashAlgorithm
  getSigner(): Crypto.Signer
}

export interface KeyManager<T extends Key> {
  keyType: KeyType

  generate(): T
  save(key: T): string
  load(value: string): T
}
