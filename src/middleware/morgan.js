const morgan = require("morgan")
const logger = require("../logger")

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

module.exports = morganMiddleware
