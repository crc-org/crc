"use strict";
const fs = require("fs");
const child_process = require("child_process");
module.exports.register = function () {
  this.on("playbookBuilt", function ({ playbook }) {
    // Get versions from Makefile
    // Use utf8 encoding to have a string rather than a buffer in the output
    const ocp_ver_full = child_process.execSync(
      "grep '^OPENSHIFT_VERSION' Makefile | cut -d' ' -f3 | tr -d '\n'",
      { encoding: "utf8" }
    );
    const ocp_ver = child_process.execSync(
      "grep '^OPENSHIFT_VERSION' Makefile | cut -d' ' -f3 | cut -d'.' -f-2 | tr -d '\n'",
      { encoding: "utf8" }
    );
    const prod_ver_full = child_process.execSync(
      "grep '^CRC_VERSION' Makefile | cut -d' ' -f3 | tr -d '\n'",
      { encoding: "utf8" }
    );
    const prod_ver = child_process.execSync(
      "grep '^CRC_VERSION' Makefile | cut -d' ' -f3 | cut -d'.' -f-2 | tr -d '\n'",
      { encoding: "utf8" }
    );
    const ushift_ver = child_process.execSync(
      "grep '^MICROSHIFT_VERSION' Makefile | cut -d' ' -f3 | tr -d '\n'",
      { encoding: "utf8" }
    );

    // Display versions
    console.log("OpenShift patch version: " + ocp_ver_full);
    console.log("OpenShift minor version: " + ocp_ver);
    console.log("CRC patch version: " + prod_ver_full);
    console.log("CRC minor version: " + prod_ver);
    console.log("MicroShift version: " + ushift_ver);

    // Set attributes values
    Object.assign(playbook.asciidoc.attributes, {
      "ocp-ver": ocp_ver,
      "ocp-ver-full": ocp_ver_full,
      "prod-ver": prod_ver,
      "prod-ver-full": prod_ver_full,
      "ushift-ver": ushift_ver,
    });
    this.updateVariables({ playbook });
  });
};
