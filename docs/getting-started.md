# Getting started with aurora

### Overiview
Aurora is a simple yet useful attempt to create a minimalistic social network with Go and
bolt database.
#### Project structure
The project is divided into two parts, a library and an app. The directories in this project
are as follows.

* bin:
This is wehere the binaries should be, unfortunate you will only find the build script file
`build.go` in this directory.
* builds:
This is where the built project is stored. Aurora needs configuration files and templates to run
so, a built binary together with all the dependency files are copied here. Inside this directory
the build versions are used to identify builds.
* config:
configurations are found here
* cmd:
This is where the aurora commandline application is.
* docs:
Project documentation
* public:
All javascript,css,fonts and images are stored here.
* templates:
templates used by aurora.


## Installation.
You will have to build this project in order to install. Make sure you have a working
golang environment, and a GOPATH.

get the project

	go get github.com/gernest/aurora


cd into the installed library

	cd $GOPATH/github.com/gernest/aurora

Run the build script(NOTE: I have used the script  for linux only, so help is needed to
provide scripts for other platforms.) I am waiting for go 1.5 to provide cross platforms
builds. This will take a while to complete as the script runs the test suite before building.

	go run bin/build.go

If you see nothing then the build was success. You should see a directory in the builds directory
with the version number e.g `0.0.1`. You can copy this folder anywhere you want or even rename it.
You can start aurora inside this directory like this

	,/aurora

A simple single command to help start aurora , you can do like this.

	cd buils/0.0.1&&./aurora           # assuming the build is version 0.0.1


After running the above command a server is started at port `8080` on localhost. So you
need to point your browser to `localhost:8080` to view the site.
