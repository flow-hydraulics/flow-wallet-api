import {createHash, Hash} from "crypto"

import {SHA3} from "sha3"

export enum HashAlgorithm {
  SHA2_256 = "SHA2_256",
  SHA3_256 = "SHA3_256",
}

export interface Hasher {
  hash(message: Buffer): Buffer
}

export class SHA2_256Hasher {
  private static shaType = "sha256"
  private sha: Hash

  constructor() {
    this.sha = createHash(SHA2_256Hasher.shaType)
  }

  hash(message: Buffer): Buffer {
    this.sha.update(message)
    return this.sha.digest()
  }
}

export class SHA3_256Hasher {
  private static size: 256 = 256
  private sha: SHA3

  constructor() {
    this.sha = new SHA3(SHA3_256Hasher.size)
  }

  hash(message: Buffer): Buffer {
    this.sha.update(message)
    return this.sha.digest()
  }
}

export function getHasher(hashAlgo: HashAlgorithm): Hasher {
  switch (hashAlgo) {
    case HashAlgorithm.SHA2_256:
      return new SHA2_256Hasher()
    case HashAlgorithm.SHA3_256:
      return new SHA3_256Hasher()
  }
}
