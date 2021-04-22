import * as express from "express"
import accounts from "./accounts"

const router = express.Router()

router.use("/accounts", accounts)

export default router
