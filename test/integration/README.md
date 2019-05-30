# CRC integration tests

CRC integration tests use Clicumber package which provides basic functionality for testing CLI binaries.
Clicumber allows to run commands in persistent shell instance (shell, tcsh, zsh, cmd, or powershell), assert its outputs (stdout, stderr or exitcode) and also allows to check configuration files etc.

The general functionality of Clicumber is then extended by CRC specific test code to cover the whole functionality of CRC.

## How to run

To start integration tests, run:

```
make integration
```

Test logs can be found in `test/integration/out/test-results`.
