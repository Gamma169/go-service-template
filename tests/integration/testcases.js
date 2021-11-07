// Configured in main.go
const DEFAULT_PORT = 7890;

function getPort() {    
  let portNum = process.env.FOOBAR_PORT || DEFAULT_PORT;
  return parseInt(portNum);
}

const SERVICE_URL = `http://127.0.0.1:${getPort()}`;

module.exports = {
  SERVICE_URL,
}
