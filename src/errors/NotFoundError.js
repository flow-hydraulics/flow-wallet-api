const { StatusCodes } = require('http-status-codes');
const ApiError = require("./ApiError");

class NotFoundError extends ApiError {
  constructor() {
    const statusCode = StatusCodes.NOT_FOUND;
    const message = "not found";

    super(statusCode, message);
  }
}

module.exports = NotFoundError;
