import * as crypto from "../crypto"

export enum KeyType {
  InMemory = "in-memory",
  GoogleKMS = "google-kms",
}

export interface Key {
  getPublicKey(): crypto.PublicKey
  getSignatureAlgorithm(): crypto.SignatureAlgorithm
  getHashAlgorithm(): crypto.HashAlgorithm
  getSigner(): crypto.Signer
}

export interface KeyManager<T extends Key> {
  keyType: KeyType

  generate(): T
  save(key: T): string
  load(value: string): T
}
