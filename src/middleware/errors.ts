import * as httpStatus from "http-status"
import * as express from "express"

import config from "src/config"
import logger from "src/logger"
import ApiError from "src/errors/ApiError"

function errorsMiddleware(
  err: ApiError,
  req: express.Request,
  res: express.Response
): void {
  let {statusCode, message} = err

  if (!(err instanceof ApiError)) {
    statusCode = httpStatus.INTERNAL_SERVER_ERROR
    message = httpStatus[httpStatus.INTERNAL_SERVER_ERROR] as string
  } else if (
    config.env === "production" &&
    statusCode === httpStatus.INTERNAL_SERVER_ERROR
  ) {
    message = httpStatus[httpStatus.INTERNAL_SERVER_ERROR] as string
  }

  const response = {
    code: statusCode,
    message,
    ...(config.env === "development" && {stack: err.stack}),
  }

  if (config.env === "development") {
    logger.error(err)
  }

  res.status(statusCode).send(response)
}

export default errorsMiddleware
