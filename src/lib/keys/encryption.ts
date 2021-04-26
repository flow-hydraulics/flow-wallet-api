import {createCipheriv, createDecipheriv, randomBytes} from "crypto"

const encryptionAlgo = "aes-256-ctr"
const ivSize = 16

export function encrypt(
  encryptionKey: Buffer,
  value: string,
): string {
  const iv = randomBytes(ivSize)

  const cipher = createCipheriv(
    encryptionAlgo, 
    encryptionKey, 
    iv,
  )

  const encrypted = cipher.update(value, "hex", "hex")

  const ivHex = iv.toString("hex")

  return ivHex + encrypted + cipher.final("hex")
}
 
export function decrypt(
  encryptionKey: Buffer,
  value: string,
): string {
  const ivHex = value.slice(0, ivSize*2)
  const ciphertextHex = value.slice(ivSize*2)

  const iv = Buffer.from(ivHex, "hex")

  const decipher = createDecipheriv(
    encryptionAlgo, 
    encryptionKey, 
    iv,
  )

  const decrypted = decipher.update(ciphertextHex, "hex", "hex")

  return decrypted + decipher.final("hex")
}
