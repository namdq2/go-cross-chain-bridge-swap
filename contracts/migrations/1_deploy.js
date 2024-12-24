const BatchBridge = artifacts.require("BatchBridge");

module.exports = function(deployer) {
    deployer.deploy(BatchBridge);
};