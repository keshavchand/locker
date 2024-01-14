# Locker 
## Getting Started

### Building Testing Binary
```bash
$ cd rootfs  
$ ./build.sh
```

### Building Engine
```bash
$ go build && ./locker
```

Note that the locker takes in an additional flag `-targetdir` where you can pass in arbitary folder. It will run the executable named `main` in the directory.
