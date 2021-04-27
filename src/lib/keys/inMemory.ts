import * as crypto from "../crypto"

import {decrypt, encrypt} from "./encryption"

import {Key, KeyManager, KeyType} from "./index"

class InMemoryKey implements Key {
  privateKey: crypto.InMemoryPrivateKey
  hashAlgo: crypto.HashAlgorithm

  constructor(
    privateKey: crypto.InMemoryPrivateKey,
    hashAlgo: crypto.HashAlgorithm
  ) {
    this.privateKey = privateKey
    this.hashAlgo = hashAlgo
  }

  getPublicKey(): crypto.PublicKey {
    return this.privateKey.getPublicKey()
  }

  getSignatureAlgorithm(): crypto.SignatureAlgorithm {
    return this.privateKey.getSignatureAlgorithm()
  }

  getHashAlgorithm(): crypto.HashAlgorithm {
    return this.hashAlgo
  }

  getSigner(): crypto.Signer {
    return new crypto.InMemorySigner(this.privateKey, this.hashAlgo)
  }
}

export default class InMemoryKeyManager implements KeyManager<InMemoryKey> {
  static keyType = KeyType.InMemory

  sigAlgo: crypto.SignatureAlgorithm
  hashAlgo: crypto.HashAlgorithm
  encryptionKey: Buffer

  constructor(
    sigAlgo: crypto.SignatureAlgorithm,
    hashAlgo: crypto.HashAlgorithm,
    encryptionKey?: Buffer
  ) {
    this.sigAlgo = sigAlgo
    this.hashAlgo = hashAlgo
    this.encryptionKey = encryptionKey
  }

  getKeyType(): KeyType {
    return InMemoryKeyManager.keyType
  }

  generate(): InMemoryKey {
    const privateKey = crypto.InMemoryPrivateKey.generate(this.sigAlgo)
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

    const privateKey = crypto.InMemoryPrivateKey.fromHex(hex, this.sigAlgo)
    return new InMemoryKey(privateKey, this.hashAlgo)
  }
}
