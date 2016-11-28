Ibex [![Build Status](https://travis-ci.org/SnapShotsApp/ibex.svg?branch=master)](https://travis-ci.org/SnapShotsApp/ibex)
====
![Ibex](logo.png)

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

Testing
-------

The test suite uses Docker to setup and populate a small test database with postgres. You can start it
with `docker-compose up`. The Rake task `rake test` waits for the database to be available before
running tests (mainly for Travis).

To run tests, you'll need to install Convey with `go get github.com/smartystreets/goconvey`. You can
then either run the suite via `go test` or run `goconvey` to bring up an autorunner in the browser.
