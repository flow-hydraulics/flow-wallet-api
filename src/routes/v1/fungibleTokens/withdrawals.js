const express = require('express');
const fungibleTokensController = require('../../../controllers/fungibleTokens');

const router = express.Router({mergeParams: true});

router
  .route('/')
  .get(fungibleTokensController.getWithdrawals)
  .post(fungibleTokensController.createWithdrawal);

router
  .route('/:transactionId')
  .get(fungibleTokensController.getWithdrawal)

module.exports = router;
