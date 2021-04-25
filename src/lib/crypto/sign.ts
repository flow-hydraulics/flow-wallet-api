export enum SignatureAlgorithm {
  ECDSA_P256 = "ECDSA_P256",
  ECDSA_secp256k1 = "ECDSA_secp256k1",
}

export interface Signature {
  toBuffer(): Buffer
  toHex(): string
}

export interface Signer {
  sign(message: Buffer): Signature
}
