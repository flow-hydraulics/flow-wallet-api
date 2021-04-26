import {AccountKey, PrismaClient} from "@prisma/client"

import {KeyType} from "src/lib/keys"

import {Account} from "./models"

export async function insertAccount(
  prisma: PrismaClient,
  address: string,
  accountKeyType: KeyType,
  accountKeyValue: string
): Promise<Account> {
  const accountKeyIndex = 0

  return await prisma.account.create({
    data: {
      address: address,
      keys: {
        create: {
          index: accountKeyIndex,
          type: accountKeyType,
          value: accountKeyValue,
        },
      },
    },
  })
}

export async function listAccounts(prisma: PrismaClient): Promise<Account[]> {
  // TODO: add pagination
  return await prisma.account.findMany()
}

export async function getAccount(
  prisma: PrismaClient,
  address: string
): Promise<Account> {
  return await prisma.account.findFirst({
    where: {address},
  })
}

export async function getAccountKey(
  prisma: PrismaClient,
  address: string
): Promise<AccountKey> {
  return await prisma.accountKey.findFirst({
    where: {accountAddress: address},
  })
}
