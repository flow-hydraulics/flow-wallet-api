require("dotenv").config()

import * as express from "express"
import morganMiddleware from "./middleware/morgan"
import errorsMiddleware from "./middleware/errors"
import NotFoundError from "./errors/NotFoundError"

import routes from "./routes/v1"

const app = express()

app.use(morganMiddleware)
app.use(express.json())
app.use(express.urlencoded({extended: false}))

app.use("/v1", routes)

// catch 404 and forward to error handler
app.use((req, res, next) => {
  next(new NotFoundError())
})

// error handler
app.use(errorsMiddleware)

export default app
