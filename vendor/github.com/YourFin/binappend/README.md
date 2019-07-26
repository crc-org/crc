# binappend
Library for packing binary data onto the end of files, targeted at adding adding static assets to executables after they're compiled.

A cli tool for reading and writing this "file format" can be found [here](https://github.com/yourfin/binappend-cli).



The following is copied from [this issue](https://github.com/gobuffalo/packr/issues/74) on github.com/gobuffalo/packr:

## Background
For one of [my own projects](https://github.com/yourfin/transcodebot), I ran into the problem that the memory overhead for embedding some of the files involved was unacceptably high over the potential lifetime of the program. I still wanted self contained binaries though, so I did some digging and ran into [this](https://stackoverflow.com/questions/5795446/appending-data-to-an-exe) on stackoverflow, and as it turns out [none of the major operating systems care if you dump crap at the end of an executable.](https://oroboro.com/packing-data-compiled-binar/) This, combined with the discovery of [the osext](https://github.com/kardianos/osext) library for finding the path to the current executable prompted the development of the following scheme for embedding files, which I have working as far as I need for my own purposes at this point.

## High level description

1. Compile the binary normally
2. Write each file to embed to the end of the binary through a `gzip.writer`, noting the start and end point on the file.
3. Write the gathered start and end points, along with their names (paths) out to a json-encoded metadata block at the end of binary
4. Write the start position of said json data out as a magic number for the last 8 bytes of the file.

The self-extraction process looks like this, then:
1. Find the path to the source of the current process
2. Open the file
3. Read the last 8 bytes
4. Open a json.Decoder at the position written to the last 8 bytes
5. Decode the json into a struct in memory
6. Close the file

Then, when the process wants to access some of the data in the file, a file reader is opened on it that is simply started at $start_ptr and is limited to reading $size bytes

## Visualization:

    0                        400                       500              7000
    +------------------------+-------------------------+----------------+---------------------+--------------+
    | Executable Binary Data | Embeded File ./assets/A | Embeded Data B | Json metadata block | Magic Number |
    +------------------------+-------------------------+----------------+---------------------+--------------+

In this case the json would look something like this:

    {
      "Version": "0.1",
      "Data":
      {
        "./assets/A": { "start_ptr": 400, "size": 100 },
        "B":          { "start_ptr": 500, "size": 6500 }
      }
    }

And the magic number would be 7000
