# aurora [![Build Status](https://drone.io/github.com/gernest/aurora/status.png)](https://drone.io/github.com/gernest/aurora/latest)[![Coverage Status](https://coveralls.io/repos/gernest/aurora/badge.svg?branch=master)](https://coveralls.io/r/gernest/aurora?branch=master)

### features
* user authentication and profile management.
* file uploads(images)
* chat( persisted in boltdb, and real time notification via websockets)


### building
* check the file `config/build/build.json` make changes if you have changed the
  project structure.
* make sure [golang](https://golang.org/) is installed, I'm currently using v1.3.3
* I assume you are running linux distro.
* Make sure [godep](https://github.com/tools/godep) is installed

* Run   `go run bin/build.go` to build.

* check built project in `builds` directory.

* cd to the builds directory and run the binary eg `cd ./builds/0.0.1&&./aurora`


## License

This project is under the MIT License. See the [LICENSE](https://github.com/gernest/nutz/blob/master/LICENCE) file for the full license text.
