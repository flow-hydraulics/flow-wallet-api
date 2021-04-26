import * as dotenv from "dotenv"
dotenv.config()

import * as express from "express"
import morganMiddleware from "./middleware/morgan"
import errorsMiddleware from "./middleware/errors"
import NotFoundError from "./errors/NotFoundError"

import createRouter from "./routes/v1"
import AccountsController from "./controllers/accounts"
import AccountsService from "./services/accounts"
import {PrismaClient} from ".prisma/client"
import FungibleTokensService from "./services/fungibleTokens"
import FungibleTokensController from "./controllers/fungibleTokens"
import InMemoryKeyManager from "./lib/keys/inMemory"
import config from "./config"
import {getAdminKey} from "./admin"

const app = express()

app.use(morganMiddleware)
app.use(express.json())
app.use(express.urlencoded({extended: false}))

const prisma = new PrismaClient()

const userKeyManager = new InMemoryKeyManager(
  config.userSigAlgo,
  config.userHashAlgo,
  config.userEncryptionKey,
)

const adminKey = getAdminKey()

const accountsService = new AccountsService(prisma, adminKey, userKeyManager)
const accountsController = new AccountsController(accountsService)

const fungiblTokensService = new FungibleTokensService(prisma, accountsService)
const fungiblTokensController = new FungibleTokensController(
  fungiblTokensService
)

const v1Router = createRouter(accountsController, fungiblTokensController)

app.use("/v1", v1Router)

// catch 404 and forward to error handler
app.use((req, res, next) => {
  next(new NotFoundError())
})

// error handler
app.use(errorsMiddleware)

export default app
