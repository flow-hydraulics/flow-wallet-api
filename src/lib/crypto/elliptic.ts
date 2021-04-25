import * as elliptic from "elliptic"
import {Endianness} from "bn.js"

import {PublicKey} from "./publicKey"
import {Signer, Signature, SignatureAlgorithm} from "./sign"
import {Hasher, HashAlgorithm, getHasher} from "./hash"

const ECDSA_P256 = new elliptic.ec("p256")
const ECDSA_secp256k1 = new elliptic.ec("secp256k1")

const bufferEndianness: Endianness = "be"

function getEC(sigAlgo: SignatureAlgorithm): elliptic.ec {
  switch (sigAlgo) {
    case SignatureAlgorithm.ECDSA_P256:
      return ECDSA_P256
    case SignatureAlgorithm.ECDSA_secp256k1:
      return ECDSA_secp256k1
  }
}

class ECSignature implements Signature {
  private static n = 32
  private ecSignature: elliptic.ec.Signature

  constructor(ecSignature: elliptic.ec.Signature) {
    this.ecSignature = ecSignature
  }

  toBuffer(): Buffer {
    const r = this.ecSignature.r.toArrayLike(
      Buffer,
      bufferEndianness,
      ECSignature.n
    )
    const s = this.ecSignature.s.toArrayLike(
      Buffer,
      bufferEndianness,
      ECSignature.n
    )

    return Buffer.concat([r, s])
  }

  toHex(): string {
    return this.toBuffer().toString("hex")
  }
}

class ECPublicKey implements PublicKey {
  private static size = 32
  private ecPublicKey: elliptic.curve.base.BasePoint

  constructor(ecPublicKey: elliptic.curve.base.BasePoint) {
    this.ecPublicKey = ecPublicKey
  }

  toBuffer(): Buffer {
    const x = this.ecPublicKey
      .getX()
      .toArrayLike(Buffer, bufferEndianness, ECPublicKey.size)
    const y = this.ecPublicKey
      .getY()
      .toArrayLike(Buffer, bufferEndianness, ECPublicKey.size)

    return Buffer.concat([x, y])
  }

  toHex(): string {
    return this.toBuffer().toString("hex")
  }
}

export class InMemoryPrivateKey {
  private ecKeyPair: elliptic.ec.KeyPair
  private sigAlgo: SignatureAlgorithm

  constructor(privateKey: Buffer, sigAlgo: SignatureAlgorithm) {
    const ec = getEC(sigAlgo)
    this.ecKeyPair = ec.keyFromPrivate(privateKey)
    this.sigAlgo = sigAlgo
  }

  public static fromHex(
    hex: string,
    sigalgo: SignatureAlgorithm
  ): InMemoryPrivateKey {
    const buffer = Buffer.from(hex, "hex")
    return new InMemoryPrivateKey(buffer, sigalgo)
  }

  sign(digest: Buffer): Signature {
    const ecSignature = this.ecKeyPair.sign(digest)
    return new ECSignature(ecSignature)
  }

  getPublicKey(): PublicKey {
    const ecPublicKey = this.ecKeyPair.getPublic()
    return new ECPublicKey(ecPublicKey)
  }

  getSignatureAlgorithm(): SignatureAlgorithm {
    return this.sigAlgo
  }

  toHex(): string {
    return this.ecKeyPair.getPrivate().toArrayLike(Buffer, "be").toString("hex")
  }
}

export class InMemorySigner implements Signer {
  private privateKey: InMemoryPrivateKey
  private hasher: Hasher

  constructor(privateKey: InMemoryPrivateKey, hashAlgo: HashAlgorithm) {
    this.privateKey = privateKey
    this.hasher = getHasher(hashAlgo)
  }

  sign(message: Buffer): Signature {
    const digest = this.hasher.hash(message)
    return this.privateKey.sign(digest)
  }
}
