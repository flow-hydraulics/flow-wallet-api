import * as morgan from "morgan"

import logger from "src/logger"

const stream = {
  write: message => logger.info(message),
}

const skip = () => {
  const env = process.env.NODE_ENV || "development"
  return env !== "development"
}

const morganMiddleware = morgan(
  ":method :url :status :res[content-length] - :response-time ms",
  {stream, skip}
)

export default morganMiddleware
