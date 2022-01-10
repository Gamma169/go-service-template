module.exports = {
  root: true,
  parserOptions: {
    ecmaVersion: 2017,
    sourceType: 'module'
  },
  rules: {
    "no-console": [1, {"allow": ["error"]}],
    "indent": [1, 2, { "SwitchCase": 1 }],
    "brace-style": [1, "stroustrup", { "allowSingleLine": true }],
    "semi": [1, "always", { "omitLastInOneLineBlock": true }],
  },
  extends: 'eslint:recommended',
  env: {
    es6: true,
    node: true,
    mocha: true
  }
};
