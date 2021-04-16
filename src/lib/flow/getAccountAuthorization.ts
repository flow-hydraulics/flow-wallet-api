import * as fcl from "@onflow/fcl"

import {Account, AccountAuthorization} from "./index"
import {SignatureAlgorithm, HashAlgorithm, signWithPrivateKey} from "./crypto"

export default function getAccountAuthorization(
  signerAddress: string,
  signerPrivateKey: string,
  signerSigAlgo: SignatureAlgorithm,
  signerHashAlgo: HashAlgorithm,
  signerKeyIndex: number
): (account?: Account) => Promise<AccountAuthorization> {
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
