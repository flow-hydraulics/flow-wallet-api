const httpStatus = require('http-status');
const config = require('../config');
const { tokens, isValidToken, getTokenTransferFunc } = require('../lib/fungibleTokens');
const getSigner = require('../lib/flow/getSigner');
const getLeastRecentAccountKey = require('../database/getLeastRecentAccountKey');
const catchAsync = require('../errors/catchAsync');
const ApiError = require('../errors/ApiError');
const InvalidFungibleTokenError = require('../errors/InvalidFungibleTokenError');

const makeToken = (tokenName) => ({ name: tokenName });
const allTokens = tokens.map((tokenName) => makeToken(tokenName));

const getTokens = catchAsync(async (req, res) => res.json(allTokens));

const getToken = catchAsync(async (req, res) => {
  const tokenName = req.params.tokenName;

  if (!isValidToken(tokenName)) {
    throw new InvalidFungibleTokenError(tokenName);
  }

  const token = makeToken(tokenName);

  res.json(token);
});

// TODO: implement withdrawal getters
const getWithdrawals = catchAsync(async (req, res) => { res.send("TODO: implement me") });
const getWithdrawal = catchAsync(async (req, res) => { res.send("TODO: implement me") });

const createWithdrawal = catchAsync(async (req, res) => {
  const tokenName = req.params.tokenName;

  if (!isValidToken(tokenName)) {
    throw new InvalidFungibleTokenError(tokenName);
  }

  // TODO: validate recipient and amount
  const { recipient, amount } = req.body;

  const signerKeyIndex = await getLeastRecentAccountKey();

  const signer = getSigner(
    config.signerAddress,
    config.signerPrivateKey,
    config.signerSigAlgo,
    config.signerHashAlgo,
    signerKeyIndex,
  )

  const transfer = getTokenTransferFunc(tokenName)
  
  try {
    const transactionId = await transfer(
      recipient,
      amount,
      signer,
      config.contracts,
    )
  
    const response = {
      transactionId,
      recipient,
      amount,
    }
  
    res.json(response);
  } catch (e) {
    throw new ApiError(
      httpStatus.INTERNAL_SERVER_ERROR,
      'failed to complete withdrawal',
    )
  }
});

module.exports = {
  getTokens,
  getToken,
  getWithdrawals,
  getWithdrawal,
  createWithdrawal,
};
