## Description

<!--
Thank you for your pull request (PR)!

Please provide a description of what your PR does providing a link (if applicable) to the issue it fixes.
-->

Fixes: #N

Relates to: #N, PR #N, ...

<!--
Describe in plain English what you solved and how. For instance, _Added `start` command to CRC so the user can create a VM and set up a single-node OpenShift cluster on it with one command. It requires blablabla..._
-->

## Type of change
<!--
What types of changes does your code introduce? Put an `x` in all the boxes that apply
-->
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] Feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change
- [ ] Chore (non-breaking change which doesn't affect codebase;
  test, version modification, documentation, etc.)

## Proposed changes
<!--
List main as well as consequential changes you introduced or had to introduce.

1. Add `start` command.
2. Add `setup` as prerequisite to `start`.
3. ...
-->

## Testing
<!--
What is the _bottom-line_ functionality that needs testing? Describe in pseudo-code or in English. Use verifiable statements that tie your changes to existing functionality.

1. `start` succeeds first time after `setup` succeeded
2. stdout contains ... if `start` succeeded
3. stderr contains ... after `start` failed
4. `status` returns ... if `start` succeeded
5. `status` returns ... if `start` failed
6. `start` fails after `start` succeeded or after `status` says CRC is `Running`
7. ...
-->

## Contribution Checklist
- [ ] I have read the [contributing guidelines](https://github.com/crc-org/crc/blob/main/developing.adoc)
- [ ] My code follows the style guidelines of this project
- [ ] I Keep It Small and Simple: The smaller the PR is, the easier it is to review and have it merged
- [ ] I have performed a self-review of my code
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I tested my code on specified platforms <!-- Only put an `x` in applicable platforms -->
    - [ ] Linux
    - [ ] Windows
    - [ ] MacOS