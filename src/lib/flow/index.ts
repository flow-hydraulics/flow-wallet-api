import getAccountAuthorization from "./getAccountAuthorization"

export interface Account {
  addr?: string
}

export interface AccountSignature {
  addr: string
  keyId: number
  signature: string
}

export interface AccountAuthorization {
  tempId: string
  addr: string
  keyId: number
  signingFunction: (data: {message: string}) => AccountSignature
}

export {getAccountAuthorization}
