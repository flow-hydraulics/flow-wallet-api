import * as express from "express"
import * as fungibleTokensController from "../../../../controllers/fungibleTokens"
import withdrawals from "./withdrawals"

const router = express.Router({mergeParams: true})

router.route("/").get(fungibleTokensController.getTokens)

router.route("/:tokenName").get(fungibleTokensController.getToken)

router.use("/:tokenName/withdrawals", withdrawals)

export default router
