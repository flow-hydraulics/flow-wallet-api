import * as express from "express"
import fungibleTokensRoute from "./fungibleTokens"

const router = express.Router()

router.use("/fungible-tokens", fungibleTokensRoute)

export default router
