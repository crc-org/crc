---
name: release-notes
description: Generate release notes for an OpenShift Local release in a text file following a pre-defined format in a text file
disable-model-invocation: true
argument-hint: [release-tag]
effort: low
---

To prepare the release notes for an OpenShift Local (crc) release:

1. Ensure we are in the crc source code repository (typically crc-org/crc)
2. Ensure the tag for the release is checked out (e.g tag name: v2.58.0) if not then abort
3. Get all the git logs between tag $0 and the last release tag (if $0 is v2.59.0 then last tag is usually v2.58.0)
    ```bash
    git log --pretty="%H %s" v2.58.0..v2.59.0
    ```
4. From each of the commit's description ignore commits that:
    - update go module dependencies
    - does not add user facing functionality or feature
    - commits that add tests
5. If the commit logs do not have the full information, track down the Pull request it belongs to based on the commit hash by running command:
    ```bash
    gh pr list --search "<COMMIT-SHA>" --state merged --json "number" --jq ".[].number"
    ```
6. With the PR number from step 5 get the PR description using command:
    ```bash 
    gh pr view <pr-number> --json "body" --jq ".body"
    ```
7. After gathering the above based on the commit log and PR description, create a descriptive management summary for each of the change and keep it under 10 to 15 words
8. Then Look at the Makefile to determine the version for the OpenShift and MicroShift bundles, they are defined as `OPENSHIFT_VERSION` and `MICROSHIFT_VERSION` in the Makefile
9. Then determine the semantic version number for the release, it is the tag or the `CRC_VERSION` variable in the Makefile, the tag adds a `v` infront of the numeric version
10. Strictly follow the template for the release notes as in [template.md](template.md) and write the new release notes to repository root to a file named in the format <crc-release-notes-2.58.0.txt>
11. Check that the generated release notes strictly resembles the expected format for the release notes as in the example file [sample.md](examples/sample.md) if it is not identical then go back to step 4 and ensure it follows the format
12. Run the [verify.py](scripts/verify.py) script with the generated file as argument to ensure all URLs are valid and return a 200 response code

