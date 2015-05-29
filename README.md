# aurora [![Build Status](https://drone.io/github.com/gernest/aurora/status.png)](https://drone.io/github.com/gernest/aurora/latest)[![Coverage Status](https://coveralls.io/repos/gernest/aurora/badge.svg?branch=master)](https://coveralls.io/r/gernest/aurora?branch=master)

# features
* user authentication and profile management.
* file uploads(images)
* chat( persisted in boltdb, and real time notification via websockets)

# building
* check the file `config/build/build.json` make changes if you have changed the
  project structure.
  
* Make sure godep is installed

* Run
    go run bin/build.go

* check your built project in `builds` directory.

    
# author
geofrey ernest

# licence
MIT