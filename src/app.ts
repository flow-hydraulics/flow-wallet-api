import * as express from "express"
import * as fcl from "@onflow/fcl"
import {PrismaClient} from "@prisma/client"

import morganMiddleware from "src/middleware/morgan"
import errorsMiddleware from "src/middleware/errors"
import NotFoundError from "src/errors/NotFoundError"
import createRouter from "src/routes/v1"
import AccountsController from "src/controllers/accounts"
import AccountsService from "src/services/accounts"
import FungibleTokensService from "src/services/fungibleTokens"
import FungibleTokensController from "src/controllers/fungibleTokens"
import TransactionsController from "src/controllers/transactions"
import TransactionsService from "src/services/transactions"
import InMemoryKeyManager from "src/lib/keys/inMemory"
import config from "src/config"
import {getAdminKey} from "src/admin"

const app = express()

app.use(morganMiddleware)
app.use(express.json())
app.use(express.urlencoded({extended: false}))

fcl.config().put("accessNode.api", config.accessApiHost)

const prisma = new PrismaClient()

const userKeyManager = new InMemoryKeyManager(
  config.userSigAlgo,
  config.userHashAlgo,
  config.userEncryptionKey
)

const adminKey = getAdminKey()

const accountsService = new AccountsService(prisma, adminKey, userKeyManager)
const accountsController = new AccountsController(accountsService)

const transactionsService = new TransactionsService(prisma, accountsService)
const transactionsController = new TransactionsController(transactionsService)

const fungibleTokensService = new FungibleTokensService(prisma, accountsService)
const fungibleTokensController = new FungibleTokensController(
  fungibleTokensService
)

const v1Router = createRouter(
  accountsController,
  transactionsController,
  fungibleTokensController
)

app.use("/v1", v1Router)

// catch 404 and forward to error handler
app.use((req, res, next) => {
  next(new NotFoundError())
})

// error handler
app.use(errorsMiddleware)

export default app
