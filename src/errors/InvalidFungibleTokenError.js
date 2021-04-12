const { StatusCodes } = require('http-status-codes');
const ApiError = require("./ApiError");

class InvalidFungibleTokenError extends ApiError {
  constructor(token) {
    const statusCode = StatusCodes.BAD_REQUEST;
    const message = `${token} is not valid fungible token`;

    super(statusCode, message);
  }
}

module.exports = InvalidFungibleTokenError;
