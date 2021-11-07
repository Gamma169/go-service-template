'use strict';
/*eslint no-console: 0 */
// shamelessly adapted from https://github.com/ember-cli/ember-cli/blob/master/tests/runner.js

const captureExit = require('capture-exit');
captureExit.captureExit();

const glob = require('glob');
const Mocha = require('mocha');

if (process.env.EOLNEWLINE) {
  require('os').EOL = '\n';
}

// Get all the test files and make sure that lint tests are run first by moving them to the front of the list
let root = '.';
let testFiles = glob.sync(`${root}/**/*-test.js`);
let lintPosition = testFiles.indexOf('./lint-test.js');
let lint = testFiles.splice(lintPosition, lintPosition+1);
testFiles = lint.concat(testFiles);



let mocha = new Mocha({
  timeout: 5000,
  slow: 3000,
  reporter: 'spec',
  retries: 0,
});
testFiles.forEach(mocha.addFile.bind(mocha));



console.time('Mocha Tests Running Time');
mocha.run(failures => {
  console.timeEnd('Mocha Tests Running Time');
  process.exit(failures);
});