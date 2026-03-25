---
name: release-notes
description: Generate release notes for an OpenShift Local release in a text file following a pre-defined format
disable-model-invocation: true
argument-hint: [release-tag]
effort: low
---

To prepare the release notes for an OpenShift Local (crc) release:

## Pre-flight Checks

1. Verify we are in the crc source code repository by checking:
    ```bash
    grep -q "github.com/crc-org/crc" go.mod
    ```
2. Verify `gh` CLI is available:
    ```bash
    command -v gh >/dev/null || { echo "ERROR: gh CLI not installed" >&2; exit 1; }
    ```
3. Ensure the tag for the release is checked out (e.g., tag name: v2.58.0):
    ```bash
    git describe --exact-match --tags
    ```
    If not on a release tag, abort with an error message.

## Gather Release Information

4. Determine the current release tag and automatically find the previous tag:
    ```bash
    CURRENT_TAG=$(git describe --exact-match --tags)
    PREV_TAG=$(git describe --tags --abbrev=0 ${CURRENT_TAG}^)
    ```

5. Get all git logs between the previous tag and current tag:
    ```bash
    git log --pretty="%H %s" ${PREV_TAG}..${CURRENT_TAG}
    ```

6. From the commit descriptions, **ignore** commits that match these patterns:
    - `build(deps):` - dependency updates
    - `chore(deps):` - dependency updates
    - `test:` or `^.*test.*$` - test-only changes
    - `ci:` - CI/CD configuration changes
    - `docs:` - documentation-only changes
    - Any commit message indicating non-user-facing changes
    
    **Keep** commits that are user-facing, such as:
    - `feat:` - new features
    - `fix:` - bug fixes
    - `perf:` - performance improvements
    - `refactor:` - code refactoring with user impact

7. For each relevant commit, get additional context from the associated Pull Request:
    ```bash
    PR_NUMBER=$(gh pr list --search "<COMMIT-SHA>" --state merged --json "number" --jq ".[].number")
    ```
    If no PR is found, try searching by commit message keywords or check if the commit was part of a squashed merge.

8. Get PR details including title, body, and labels:
    ```bash 
    gh pr view ${PR_NUMBER} --json "title,body,labels" --jq '{title, body, labels: [.labels[].name]}'
    ```
    Use labels to identify breaking changes, features, or bug fixes if available.

## Extract Version Information

9. Extract the OpenShift and MicroShift bundle versions from the Makefile:
    ```bash
    OPENSHIFT_VERSION=$(grep "^OPENSHIFT_VERSION" Makefile | cut -d'=' -f2 | tr -d ' ')
    MICROSHIFT_VERSION=$(grep "^MICROSHIFT_VERSION" Makefile | cut -d'=' -f2 | tr -d ' ')
    ```

10. Determine the semantic version number for the release:
    ```bash
    CRC_VERSION=${CURRENT_TAG#v}  # Remove 'v' prefix from tag
    ```

11. Validate all required template variables are available:
    - `${CRC_VERSION}` - numeric version (e.g., 2.58.0)
    - `${CURRENT_TAG}` - git tag with 'v' prefix (e.g., v2.58.0)
    - `${OPENSHIFT_VERSION}` - OpenShift bundle version
    - `${MICROSHIFT_VERSION}` - MicroShift bundle version

## Generate Release Notes

12. Create descriptive summaries for each user-facing change:
    - Keep each summary between 10-15 words
    - Focus on user impact, not implementation details
    - Reference the issue number or PR number in brackets (e.g., `[4]`, `[5]`)

13. Strictly follow the template format from [template.md](template.md) and write the release notes to:
    ```
    Output file: crc-release-notes-${CRC_VERSION}.txt
    Example: crc-release-notes-2.58.0.txt
    ```
    Place the file in the repository root directory.

## Validation

14. Verify the generated release notes match the expected format in [sample.md](examples/sample.md):
    - Check Subject line format
    - Verify all version placeholders are replaced
    - Ensure bullet points are properly formatted
    - Confirm all reference links `[0]` through `[n]` are present and sequential

15. Verify all URLs are valid and return HTTP 200:
    ```bash
    cat crc-release-notes-${CRC_VERSION}.txt | python3 .claude/skills/release-notes/scripts/verify.py
    ```

## Post-Generation Summary

16. Provide a summary including:
    - Total number of commits reviewed
    - Number of user-facing changes included in release notes
    - List of PRs/issues referenced
    - Confirmation that all URLs are valid
    - Output file location
    - Next steps: "Review the release notes and send via email"

