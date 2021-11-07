'use strict';

const glob = require('glob').sync;

let paths = glob('*');

const filesAndDirectoriesToIgnore = ['node_modules', 'yarn.lock', 'package.json'];

paths = paths.filter(file => !filesAndDirectoriesToIgnore.includes(file));

require('mocha-eslint')(paths, {
  timeout: 1000,
  slow: 500,
  retries: 0,
});
