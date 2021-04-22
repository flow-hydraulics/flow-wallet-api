import * as fcl from "@onflow/fcl"
import * as dotenv from "dotenv"

dotenv.config()

import {PrismaClient} from "@prisma/client"

const prisma = new PrismaClient()

const adminAddress = process.env.ADMIN_ADDRESS
const accessAPIHost = process.env.ACCESS_API_HOST

fcl.config().put("accessNode.api", accessAPIHost)

async function getAccount(address) {
  const {account} = await fcl.send([fcl.getAccount(address)])
  return account
}

async function main() {
  const account = await getAccount(adminAddress)

  console.log(
    `Fetched admin account information for ${adminAddress} from ${accessAPIHost}\n`
  )

  // truncate existing keys
  const {count} = await prisma.adminSignerKey.deleteMany({where: {}})

  console.log(`Removed ${count} existing admin key(s) from DB\n`)

  await Promise.all(
    account.keys.map(async key => {
      console.log(`- Inserting key ${key.index}`)

      await prisma.adminSignerKey.create({
        data: {index: key.index},
      })
    })
  )

  console.log(`\nInserted ${account.keys.length} admin key(s) into DB`)
}

main()
  .catch(e => {
    throw e
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
