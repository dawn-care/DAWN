const path = require("path");
const { getDefaultConfig } = require("expo/metro-config");

const config = getDefaultConfig(__dirname);

// `shared/` sits outside this app, so Metro has to watch it as well.
config.watchFolders = [...config.watchFolders, path.resolve(__dirname, "../shared")];

module.exports = config;
