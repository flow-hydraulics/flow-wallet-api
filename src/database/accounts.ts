import {Prisma, PrismaClient} from "@prisma/client"

export async function insertAccount(
  prisma: PrismaClient,
  account: Prisma.AccountCreateInput
) {
  await prisma.account.create({
    data: {
      address: account.address,
    },
  })
}
