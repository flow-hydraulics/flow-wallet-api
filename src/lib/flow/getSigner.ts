import * as fcl from "@onflow/fcl"

import {signWithPrivateKey} from "./crypto"

export default function getSigner(
  signerAddress,
  signerPrivateKey,
  signerSigAlgo,
  signerHashAlgo,
  signerKeyIndex
) {
  return async (account = {}) => {
    return {
      ...account,
      tempId: "SIGNER",
      addr: fcl.sansPrefix(signerAddress),
      keyId: signerKeyIndex,
      signingFunction: data => ({
        addr: fcl.withPrefix(signerAddress),
        keyId: signerKeyIndex,
        signature: signWithPrivateKey(
          signerPrivateKey,
          signerSigAlgo,
          signerHashAlgo,
          data.message
        ),
      }),
    }
  }
}
