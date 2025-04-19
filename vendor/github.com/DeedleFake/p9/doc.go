// Package p9 contains an implementation of 9P, the Plan 9 from Bell
// Labs file protocol.
//
// The package provides high-level APIs for both creating 9P servers
// and connecting to those servers as a client. Although it abstracts
// away a lot of the complexity of 9P, some familiarity with the
// protocol is advised. Like "net/http", it exposes a fair amount of
// the inner workings of the package so that a user can opt to build
// their own high-level implementation on top of it.
//
// The primary concept that the user should understand is that in 9P,
// everything is referenced relative to previously referenced objects.
// For example, a typical 9P exchange, skipping version negotiation
// and authentication, might look something like this:
//
//	attach to filesystem "/" and call it 1
//	navigate to file at "some/path" relative to 1 and call it 2
//	navigate to file at "../woops/other/path" relative to 2 and call it 3
//	open file 3
//	read from file 3
//
// This package attempts to completely abstract away the navigation
// aspects of 9P, but a lot of things are still relative to others.
// For example, opening a file on the server from the client is done
// by calling the Open method on an already existing file reference
// and passing it a path.
//
// The Client type provides a series of functionality that allows the
// user to connect to 9P servers. Here's an example of its use:
//
//	c, _ := p9.Dial("tcp", addr)
//	defer c.Close()
//	c.Handshake(4096)
//
//	root, _ := c.Attach(nil, "anyone", "/")
//	defer root.Close()
//
//	file, _ := root.Open("path/to/a/file", p9.OREAD)
//	defer file.Close()
//	buf, _ := ioutil.ReadAll(file)
//
// The client is split into two main types: Client and Remote. Client
// provides the basic functionality for establishing a connection,
// performing authentication, and attaching to file hierarchies.
// Remote provides functionality for opening and creating files,
// getting information about them, and reading from and writing to
// them. They behave similarly to files themselves, implementing many
// of the same interfaces that os.File implements.
//
// The server works similarly to the "net/http" package, but, due to
// the major differences in the protocol being handled, is quite a bit
// more complicated. At the top level, there is a ListenAndServe
// function, much like what is provided by "net/http". For most cases,
// the user can provide a filesystem by implementing the FileSystem
// interface and passing an instance of their implementation to the
// FSConnHandler function to get a handler to pass to ListenAndServe.
//
// If the user only wants to serve local files, the Dir type provides
// a pre-built implementation of FileSystem that does just that.
// Similarly, the AuthFS type allows the user to add the ability to
// authenticate to a FileSystem implementation that otherwise has
// none, such as the aforementioned Dir.
package p9
