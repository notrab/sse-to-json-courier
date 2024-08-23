const { Binary } = require("binary-install");
const os = require("os");
const packageJson = require("../package.json");

function getBinary() {
  const type = os.type();
  const arch = os.arch();

  if (type === "Darwin") {
    if (arch === "arm64") {
      return "sse-proxy-cli-darwin-arm64";
    }
    return "sse-proxy-cli-darwin-x64";
  }
  if (type === "Linux") {
    if (arch === "arm64") {
      return "sse-proxy-cli-linux-arm64";
    }
    return "sse-proxy-cli-linux-x64";
  }
  throw new Error(`Unsupported platform: ${type} ${arch}`);
}

function getBinaryUrl() {
  const version = packageJson.version;
  const url = `https://github.com/notrab/sse-to-json-courier/releases/download/v${version}/${getBinary()}.tar.gz`;
  return url;
}

function getPlatformBinary() {
  const binary = getBinary();
  const url = getBinaryUrl();
  return new Binary(binary, url);
}

module.exports = {
  getBinary,
  getBinaryUrl,
  getPlatformBinary,
};
