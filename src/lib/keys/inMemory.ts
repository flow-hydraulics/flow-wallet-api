import * as Crypto from "../crypto"
import { decrypt, encrypt } from "./encryption"
import {Key, KeyManager, KeyType} from "./index"

class InMemoryKey implements Key {
  privateKey: Crypto.InMemoryPrivateKey
  hashAlgo: Crypto.HashAlgorithm

  constructor(
    privateKey: Crypto.InMemoryPrivateKey,
    hashAlgo: Crypto.HashAlgorithm
  ) {
    this.privateKey = privateKey
    this.hashAlgo = hashAlgo
  }

  getPublicKey(): Crypto.PublicKey {
    return this.privateKey.getPublicKey()
  }

  getSignatureAlgorithm(): Crypto.SignatureAlgorithm {
    return this.privateKey.getSignatureAlgorithm()
  }

  getHashAlgorithm(): Crypto.HashAlgorithm {
    return this.hashAlgo
  }

  getSigner(): Crypto.Signer {
    return new Crypto.InMemorySigner(this.privateKey, this.hashAlgo)
  }
}

export default class InMemoryKeyManager implements KeyManager<InMemoryKey> {
  keyType = KeyType.InMemory
  sigAlgo: Crypto.SignatureAlgorithm
  hashAlgo: Crypto.HashAlgorithm
  encryptionKey: Buffer

  constructor(
    sigAlgo: Crypto.SignatureAlgorithm,
    hashAlgo: Crypto.HashAlgorithm,
    encryptionKey?: Buffer,
  ) {
    this.sigAlgo = sigAlgo
    this.hashAlgo = hashAlgo
    this.encryptionKey = encryptionKey
  }

  generate(): InMemoryKey {
    const privateKey = Crypto.InMemoryPrivateKey.generate(this.sigAlgo)
    return new InMemoryKey(privateKey, this.hashAlgo)
  }

  save(key: InMemoryKey): string {
    const hex = key.privateKey.toHex()

    if (this.encryptionKey) {
      return encrypt(this.encryptionKey, hex)
    }

    return hex
  }

  load(value: string): InMemoryKey {
    let hex = value

    if (this.encryptionKey) {
      hex = decrypt(this.encryptionKey, hex)
    }

    const privateKey = Crypto.InMemoryPrivateKey.fromHex(hex, this.sigAlgo)
    return new InMemoryKey(privateKey, this.hashAlgo)
  }
}
