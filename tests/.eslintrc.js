module.exports = {
  root: true,
  parserOptions: {
    ecmaVersion: 2017,
    sourceType: 'module'
  },
  rules: {
    "no-console": [2, {"allow": ["error"]}],
    "indent": [1, 2, { "SwitchCase": 1 }],
    "brace-style": [1, "stroustrup", { "allowSingleLine": true }],
    "semi": [1, "always", { "omitLastInOneLineBlock": true }],
    "no-trailing-spaces": 1,
  },
  extends: 'eslint:recommended',
  env: {
    es6: true,
    node: true,
    mocha: true
  }
};
