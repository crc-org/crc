# go9p
This is an implementation of the 9p2000 protocol in Go.

There are a couple of packages here.

[`github.com/knusbaum/go9p`](http://godoc.org/github.com/knusbaum/go9p) just contains an interface definition for a 9p2000 server, `Srv`.
along with a few functions that will serve the 9p2000 protocol using a `Srv`.

[`github.com/knusbaum/go9p/proto`](http://godoc.org/github.com/knusbaum/go9p/proto) is the protocol implementation. It is used by the other packages to
send and receive 9p2000 messages. It may be useful to someone who wants to investigate 9p2000 at the
protocol level.

[`github.com/knusbaum/go9p/fs`](http://godoc.org/github.com/knusbaum/go9p/fs) is an package that implements a hierarchical filesystem as a struct, `FS`.
An `FS` contains a hierarchy of `Dir`s and `File`s. The package also contains other types and functions 
useful for building 9p filesystems.

Examples are available in examples/

This library has been tested on Plan 9 and Linux.

Programs serving 9p2000 can be mounted on Unix with plan9port's 9pfuse (or some equivalent)
https://github.com/9fans/plan9port

This repository now also offers the [mount9p](cmd/mount9p) and [export9p](cmd/export9p) programs.
mount9p replaces plan9port's 9pfuse and export9p will export part of a local namespace via 9p.

For example, you would mount the ramfs example with the following command:
```
9pfuse localhost:9999 /mnt/myramfs
```
Then you can copy files to/from the ramfs and do all the other stuff that you'd expect.

This is distributed under the MIT license

```
    Copyright (c) 2016-Present Kyle Nusbaum


    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is furnished
    to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all
    copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
    FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
    COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
    AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
    WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```