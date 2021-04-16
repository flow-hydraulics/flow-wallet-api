import * as httpStatus from "http-status"
import ApiError from "./ApiError"

export default class InvalidFungibleTokenError extends ApiError {
  constructor(token: string) {
    const statusCode = httpStatus.BAD_REQUEST
    const message = `${token} is not valid fungible token`

    super(statusCode, message)
  }
}
