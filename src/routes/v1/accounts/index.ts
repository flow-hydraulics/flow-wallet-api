import * as express from "express"
import * as accountsController from "../../../controllers/accounts"
import fungibleTokens from "./fungibleTokens"

const router = express.Router()

router.route("/").get(accountsController.getAccounts)

router.route("/:address").get(accountsController.getAccount)

router.use("/:address/fungible-tokens", fungibleTokens)

export default router
