# aurora

Messaging app based on go and boltdb. 


# how to build

check the file `config/build/build.json` make changes if you have changed the
 project structure.

You only need golang environmant to build(go 1.3+ )

 Clone this repo in your GOPATH

    git clone

Now cd into your cloned repo and execute the following command. Note that you
 must have all the go tools command in your system path.  The build script
 uses go toolsets, you can check yourself by opening the file `bin/build.go`.

    go run bin/build.go

if you see nothing is printed then your build was okay. and the built app
will be inside the directory `builds/{ version number}/` where version number
 is the one specified in `config/build/build.json`

# Run the app
agter building you can cd into the build directory where the execurable
aurora resides and do this.

    ./aurora



