
**Fixes:** Issue #N

**Relates to:** Issue #1, Issue #2, PR #1, ...

## Solution/Idea

Describe in plain English what you solved and how. For instance, _Added `start` command to CRC so the user can create a VM and set-up a single-node OpenShift cluster on it with one command. It required blablabla..._

## Proposed changes

List main as well as consequential changes you introduced or had to introduce.

1. Add `start` command.
2. Add `setup` as prerequisite to `start`.
3. ...

## Testing

What is the _bottom-line_ functionality that needs testing? Describe in pseudo-code or in English. Use verifiable statements that tie your changes to existing functionality.

1. `start` succeeds first time after `setup` succeeded
2. stdout contains ... if `start` succeeded
3. stderr contains ... after `start` failed
4. `status` returns ... if `start` succeeded
5. `status` returns ... if `start` failed
6. `start` fails after `start` succeeded or after `status` says CRC is `Running`
7. ...
