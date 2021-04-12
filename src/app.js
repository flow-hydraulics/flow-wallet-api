require('dotenv').config()

const express = require('express');
const morganMiddleware = require('./middleware/morgan');
const errorsMiddleware = require('./middleware/errors');
const NotFoundError = require('./errors/NotFoundError');

const routes = require('./routes/v1');

const app = express();

app.use(morganMiddleware);
app.use(express.json());
app.use(express.urlencoded({ extended: false }));

app.use('/v1', routes);

// catch 404 and forward to error handler
app.use(function(req, res, next) {
  next(new NotFoundError());
});

// error handler
app.use(errorsMiddleware);

module.exports = app;
