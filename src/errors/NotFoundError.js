const httpStatus = require("http-status")
const ApiError = require("./ApiError")

class NotFoundError extends ApiError {
  constructor() {
    const statusCode = httpStatus.NOT_FOUND
    const message = "not found"

    super(statusCode, message)
  }
}

module.exports = NotFoundError
