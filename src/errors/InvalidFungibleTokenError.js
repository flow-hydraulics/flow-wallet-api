const httpStatus = require('http-status');
const ApiError = require("./ApiError");

class InvalidFungibleTokenError extends ApiError {
  constructor(token) {
    const statusCode = httpStatus.BAD_REQUEST;
    const message = `${token} is not valid fungible token`;

    super(statusCode, message);
  }
}

module.exports = InvalidFungibleTokenError;
