# go-fuse-sample

## Developed by Mac

1. Virtual Box
2. Vagrant

```sh
$ brew cask install virtualbox
$ brew cask install vagrant
$ vagrant -v
$ vagrant plugin install vagrant-disksize vagrant-hostsupdater vagrant-mutagen
```

## 

```sh
$ docker run --rm -it -v $(pwd):/go/src/github.com/transnano/go-fuse-sample/ -w /go/src/github.com/transnano/go-fuse-sample golang:1.14.4 bash

docker run --rm -it -v $(pwd):/go/src/github.com/transnano/go-fuse-sample/ -w /go/src/github.com/transnano/go-fuse-sample golang:1.14.4 go build -o hello main.go

$ uname -a
Linux 2edb5fdeab70 4.19.76-linuxkit #1 SMP Tue May 26 11:42:35 UTC 2020 x86_64 GNU/Linux
$ uname -s
Linux
$ uname -m
x86_64
$ go env GOOS
linux
$ go env GOARCH
amd64
$ go build -o hello main.go

$ docker run --rm -it ubuntu:18.04 bash
fuse
```

## check

```sh
docker build -t go-fuse-sample:latest -f ./Dockerfile.check .
docker run -it --rm --device /dev/fuse --cap-add SYS_ADMIN go-fuse-sample:latest bash


docker run -it --rm -v $(pwd):/app -w /app --device /dev/fuse --cap-add SYS_ADMIN go-fuse-sample:latest bash

root@ce6476928df1:/app# mkdir /tmp/mountpoint
root@ce6476928df1:/app# fusermount -u /tmp/mountpoint
root@ce6476928df1:/app# ./hello /tmp/mountpoint &
ls /tmp/mountpoint

docker run -d --rm \
           --device /dev/fuse \
           --cap-add SYS_ADMIN \
      go-fuse-sample:latest bash

docker run -d --rm \
           --device /dev/fuse \
           --cap-add SYS_ADMIN \
           --security-opt apparmor:unconfined \
      <image_id/name>
```
