const express = require('express');
const axios = require('axios');
const _ = require('lodash');

const app = express();

app.get('/', (req, res) => {
  const data = _.merge({}, req.query);
  res.json(data);
});

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
