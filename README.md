Ibex
====
A custom reverse-proxy for Imagizer.

Building
--------
You'll need to have Go on your system. Something like `brew install go` should do.

You'll also need `go-bindata` to package up the resources: `go get -u github.com/jteeuwen/go-bindata/...`

For dependency management, we use `godep`: `go get github.com/tools/godep`. And then run `godep get`
in the repo directory.

Then just run `rake`.

Developing
----------

You want to regenerate the bindata in debug mode. This loads it from disk instead of embedding it in the
`bindata.go` file: `go-bindata -debug resources/`.

Then you can start the server with a command like `go run *.go resources/config.json`
