import * as httpStatus from "http-status"
import config from "../config"
import logger from "../logger"
import ApiError from "../errors/ApiError"

const errorsMiddleware = (err, req, res, next) => {
  let {statusCode, message} = err

  if (!(err instanceof ApiError)) {
    statusCode = httpStatus.INTERNAL_SERVER_ERROR
    message = httpStatus[httpStatus.INTERNAL_SERVER_ERROR]
  } else if (
    config.env === "production" &&
    statusCode === httpStatus.INTERNAL_SERVER_ERROR
  ) {
    message = httpStatus[httpStatus.INTERNAL_SERVER_ERROR]
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
