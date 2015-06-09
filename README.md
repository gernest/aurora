# aurora [![Build Status](https://drone.io/github.com/gernest/aurora/status.png)](https://drone.io/github.com/gernest/aurora/latest)[![Coverage Status](https://coveralls.io/repos/gernest/aurora/badge.svg?branch=master)](https://coveralls.io/r/gernest/aurora?branch=master)

## Guide
1. Getting started
	- Overview
	- features
	- Installation
2. Routes
	- Overview
	- Home
	- Register
	- Login
	- ServeImages
	- Uploads
	- Logout
	- Profile
3. Chat
	- Overview
	- Server side
	- Client side
4. Forms
5. COnfiguration
6. Sessions
7. File uploads.









### features
* user authentication and profile management.
* file uploads(images)
* chat( persisted in boltdb, and real time notification via websockets)


### building

* make sure [golang](https://golang.org/) is installed, I'm currently using v1.3.3

* go get the project.

		go get github.com/gernest/aurora

* cd to the project path

		cd $GOPATH/github.com/gernest/aurora


* Run   `go run bin/build.go` to build.

* check built project in `builds` directory inside the project path.

* cd to the builds directory and run the binary eg `cd ./builds/0.0.1&&./aurora`


## License

This project is under the MIT License. See the [LICENSE](https://github.com/gernest/aurora/blob/master/LICENCE) file for the full license text.
