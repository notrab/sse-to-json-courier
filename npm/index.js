const { spawn } = require("child_process");
const { getPlatformBinary } = require("./binary");

const runSSEProxy = (options) => {
  return new Promise((resolve, reject) => {
    const binary = getPlatformBinary();
    const args = ["start"];
    if (options.sourceUrl) args.push(`--source=${options.sourceUrl}`);
    if (options.targetUrl) args.push(`--target=${options.targetUrl}`);
    if (options.authToken) args.push(`--auth=${options.authToken}`);
    if (options.port) args.push(`--port=${options.port}`);

    const child = spawn(binary.binaryPath, args, { stdio: "inherit" });

    child.on("error", (err) => {
      reject(err);
    });

    child.on("exit", (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`Process exited with code ${code}`));
      }
    });
  });
};

module.exports = runSSEProxy;
