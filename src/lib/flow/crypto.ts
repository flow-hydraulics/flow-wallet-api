import {createHash} from "crypto"
import {SHA3} from "sha3"
import {ec as EC} from "elliptic"

export enum SignatureAlgorithm {
  ECDSA_P256 = "ECDSA_P256",
  ECDSA_secp256k1 = "ECDSA_secp256k1",
}

export enum HashAlgorithm {
  SHA2_256 = "SHA2_256",
  SHA3_256 = "SHA3_256",
}

const signers = {
  ECDSA_P256: () => new EC("p256"),
  ECDSA_secp256k1: () => new EC("secp256k1"),
}

function hashSHA2(msg) {
  const sha = createHash("sha256")
  sha.update(Buffer.from(msg, "hex"))
  return sha.digest()
}

function hashSHA3(msg) {
  const sha = new SHA3(256)
  sha.update(Buffer.from(msg, "hex"))
  return sha.digest()
}

const hashers = {
  SHA2_256: hashSHA2,
  SHA3_256: hashSHA3,
}

const getSigner = sigAlgo => signers[sigAlgo]()
const getHasher = hashAlgo => hashers[hashAlgo]

function encodeSignature(rawSig): string {
  const n = 32
  const r = rawSig.r.toArrayLike(Buffer, "be", n)
  const s = rawSig.s.toArrayLike(Buffer, "be", n)

  return Buffer.concat([r, s]).toString("hex")
}

export function signWithPrivateKey(
  privateKey: string,
  sigAlgo: SignatureAlgorithm,
  hashAlgo: HashAlgorithm,
  msg: string
): string {
  const signer = getSigner(sigAlgo)
  const hasher = getHasher(hashAlgo)

  const key = signer.keyFromPrivate(Buffer.from(privateKey, "hex"))
  const digest = hasher(msg)

  const sig = key.sign(digest)

  return encodeSignature(sig)
}
