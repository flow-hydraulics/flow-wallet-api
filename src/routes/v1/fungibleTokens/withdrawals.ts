import * as express from "express"
import * as fungibleTokensController from "../../../controllers/fungibleTokens"

const router = express.Router({mergeParams: true})

router
  .route("/")
  .get(fungibleTokensController.getWithdrawals)
  .post(fungibleTokensController.createWithdrawal)

router.route("/:transactionId").get(fungibleTokensController.getWithdrawal)

export default router
