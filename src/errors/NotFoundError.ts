import * as httpStatus from "http-status"

import ApiError from "./ApiError"

export default class NotFoundError extends ApiError {
  constructor() {
    const statusCode = httpStatus.NOT_FOUND
    const message = "not found"

    super(statusCode, message)
  }
}
