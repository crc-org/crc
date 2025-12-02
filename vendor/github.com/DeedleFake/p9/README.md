p9
==

[![Go Report Card](https://www.goreportcard.com/badge/github.com/DeedleFake/p9)](https://www.goreportcard.com/report/github.com/DeedleFake/p9)
[![GoDoc](http://www.godoc.org/github.com/DeedleFake/p9?status.svg)](http://www.godoc.org/github.com/DeedleFake/p9)

An experimental Go package for dealing with 9P, the Plan 9 file protocol. The primary idea behind this package is to make building 9P servers and clients as simple as building HTTP servers and clients is in Go. Due to the complexity of the protocol compared to HTTP, this package is unlikely to reach that level of simplicity, but it's certainly simpler than many other existing packages.

Example
-------

### Server

```go
err := p9.ListenAndServe("tcp", "localhost:5640", p9.FSConnHandler(fsImplementation))
if err != nil {
	log.Fatalf("Failed to start server: %v", err)
}
```

### Client

```go
c, err := p9.Dial("tcp", "localhost:5640")
if err != nil {
	log.Fatalf("Failed to dial address: %v", err)
}
defer c.Close()

_, err := c.Handshake(2048)
if err != nil {
	log.Fatalf("Failed to perform handshake: %v", err)
}

root, err := c.Attach(nil, "anyone", "/")
if err != nil {
	log.Fatalf("Failed to attach: %v", err)
}
defer root.Close()

file, err := root.Open("path/to/a/file.txt", p9.OREAD)
if err != nil {
	log.Fatalf("Failed to open file: %v", err)
}
defer file.Close()

_, err = io.Copy(os.Stdout, file)
if err != nil {
	log.Fatalf("Failed to read file: %v", err)
}
```
