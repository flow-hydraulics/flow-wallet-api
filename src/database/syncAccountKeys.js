const fcl = require("@onflow/fcl")
const dotenv = require("dotenv")

dotenv.config()

const {PrismaClient} = require("@prisma/client")

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
    `Fetched account information for ${adminAddress} from ${accessAPIHost}\n`
  )

  // truncate existing keys
  const {count} = await prisma.accountKey.deleteMany({where: {}})

  console.log(`Removed ${count} existing key(s) from DB\n`)

  await Promise.all(
    account.keys.map(async key => {
      console.log(`- Inserting key ${key.index}`)

      await prisma.accountKey.create({
        data: {index: key.index},
      })
    })
  )

  console.log(`\nInserted ${account.keys.length} key(s) into DB`)
}

main()
  .catch(e => {
    throw e
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
