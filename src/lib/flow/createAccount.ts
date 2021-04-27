import * as dedent from "dedent-js"
import {
  ECDSA_P256,
  ECDSA_secp256k1,
  SHA2_256,
  SHA3_256,
  encodeKey,
} from "@onflow/util-encode-key"

import * as crypto from "../crypto"

import sendTransaction from "./sendTransaction"

import {AccountAuthorizer} from "./index"

const sigAlgos = {
  [crypto.SignatureAlgorithm.ECDSA_P256]: ECDSA_P256,
  [crypto.SignatureAlgorithm.ECDSA_secp256k1]: ECDSA_secp256k1,
}

const hashAlgos = {
  [crypto.HashAlgorithm.SHA2_256]: SHA2_256,
  [crypto.HashAlgorithm.SHA3_256]: SHA3_256,
}

const accountKeyWeight = 1000

function txCreateAccount(contracts) {
  return dedent`
  import FungibleToken from ${contracts.FungibleToken}
  import FUSD from ${contracts.FUSD}

  transaction(publicKey: String) {

    let account: AuthAccount

    prepare(signer: AuthAccount) {
      self.account = AuthAccount(payer: signer)
    }

    execute {
      self.account.addPublicKey(publicKey.decodeHex())

      // Add FUSD vault
      self.account.save(<-FUSD.createEmptyVault(), to: /storage/fusdVault)

      self.account.link<&FUSD.Vault{FungibleToken.Receiver}>(
          /public/fusdReceiver,
          target: /storage/fusdVault
      )

      self.account.link<&FUSD.Vault{FungibleToken.Balance}>(
          /public/fusdBalance,
          target: /storage/fusdVault
      )
    }
  }
  `
}

export async function createAccount(
  publicKey: crypto.PublicKey,
  sigAlgo: crypto.SignatureAlgorithm,
  hashAlgo: crypto.HashAlgorithm,
  authorization: AccountAuthorizer,
  contracts: {[key: string]: string}
): Promise<string> {
  const encodedPublicKey = encodeKey(
    publicKey.toHex(),
    sigAlgos[sigAlgo],
    hashAlgos[hashAlgo],
    accountKeyWeight
  )

  const result = await sendTransaction({
    code: txCreateAccount(contracts),
    args: [
      {
        type: "String",
        value: encodedPublicKey,
      },
    ],
    authorizations: [authorization],
    payer: authorization,
    proposer: authorization,
  })

  const accountCreatedEvent = result.events[0].data

  return accountCreatedEvent.address
}
