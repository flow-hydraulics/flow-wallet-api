const express = require("express")
const fungibleTokensController = require("../../../controllers/fungibleTokens")
const withdrawals = require("./withdrawals")

const router = express.Router()

router.route("/").get(fungibleTokensController.getTokens)

router.route("/:tokenName").get(fungibleTokensController.getToken)

router.use("/:tokenName/withdrawals", withdrawals)

module.exports = router
