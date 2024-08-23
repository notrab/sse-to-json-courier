#!/usr/bin/env node

const { getPlatformBinary } = require("./binary");

const binary = getPlatformBinary();
binary.install();
