const path = require("path");
const { getDefaultConfig } = require("expo/metro-config");

const config = getDefaultConfig(__dirname);

// `shared/` sits outside this app, so Metro has to watch it as well.
config.watchFolders = [...config.watchFolders, path.resolve(__dirname, "../shared")];

// Files in `shared/` are compiled like app code, so they need this app's packages
// (for example the Babel helpers) and there is no node_modules next to them.
config.resolver.nodeModulesPaths = [path.resolve(__dirname, "node_modules")];

module.exports = config;
