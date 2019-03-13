# CRC integration tests

CRC integration tests use Clicumber package which provides basic functionality for testing CLI binaries.
Clicumber allows to run commands in persistent shell instance (shell, tcsh, zsh, cmd, or powershell), assert its outputs (stdout, stderr or exitcode) and also allows to check configuration files etc.

The general functionality of Clicumber is then extended by CRC specific test code to cover the whole functionality of CRC.

## How to run

1. get the Clicumber package: `go get -u github.com/agajdosi/clicumber`
2. change the working directory to: `test/integration`
3. make sure the tested `crc` binary is on the path (support for automatic setup should come later)
4. to start the integration tests then run:

```
go test --tags integration
```

5. log of the test run can be seen at `test/integration/out/test-results`