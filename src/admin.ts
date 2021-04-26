import config from "./config"
import {KeyType, Key} from "./lib/keys"
import InMemoryKeyManager from "./lib/keys/inMemory"

export function getAdminKey(): Key {
  switch (config.adminKeyType) {
    case KeyType.InMemory:
      return getAdminInMemoryKey()
  }
}

function getAdminInMemoryKey(): Key {
  const keyManager = new InMemoryKeyManager(
    config.adminSigAlgo,
    config.adminHashAlgo,
  )

  return keyManager.load(config.adminPrivateKey)
}
