import * as fcl from "@onflow/fcl"

import {AccountAuthorizer} from "./index"
import * as Crypto from "../crypto"

const fromHex = (hex: string) => Buffer.from(hex, "hex")

export default function getAuthorization(
  address: string,
  keyIndex: number,
  signer: Crypto.Signer
): AccountAuthorizer {
  return async (account = {}) => {
    return {
      ...account,
      tempId: "SIGNER",
      addr: fcl.sansPrefix(address),
      keyId: keyIndex,
      signingFunction: data => ({
        addr: fcl.withPrefix(address),
        keyId: keyIndex,
        signature: signer.sign(fromHex(data.message)).toHex(),
      }),
    }
  }
}
