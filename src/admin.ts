import config from "src/config"
import {KeyType, Key} from "src/lib/keys"
import InMemoryKeyManager from "src/lib/keys/inMemory"

export function getAdminKey(): Key {
  switch (config.adminKeyType) {
    case KeyType.InMemory:
      return getAdminInMemoryKey()
  }
}

function getAdminInMemoryKey(): Key {
  const keyManager = new InMemoryKeyManager(
    config.adminSigAlgo,
    config.adminHashAlgo
  )

  return keyManager.load(config.adminPrivateKey)
}
