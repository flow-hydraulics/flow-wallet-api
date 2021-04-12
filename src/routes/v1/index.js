const express = require('express');
const fungibleTokensRoute = require('./fungibleTokens');

const router = express.Router();

router.use('/fungible-tokens', fungibleTokensRoute);

module.exports = router;
