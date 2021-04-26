import {Prisma, PrismaClient} from "@prisma/client"
import {Account} from "./models"

export async function insertAccount(
  prisma: PrismaClient,
  address: string
): Promise<Account> {
  return await prisma.account.create({
    data: {
      address: address,
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
