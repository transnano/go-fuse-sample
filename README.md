# go-fuse-sample

## 

```sh
# Usage:\n  hello MOUNTPOINT
$ sudo ./hello /tmp/mountpoint
$ sudo ./hello -o allow_other -c ~/.pulsar-fuse.yaml /tmp/mountpoint

$ sudo ./pulsarfs --help
Usage: pulsarfs [flags] mountpoint
  -c, --config string     config file path
  -u, --username string   mycloud.com username
  -p, --password string   mycloud.com password
  -a, --allow-other       allow other users
  -U, --uid int           set the owner of the files in the filesystem (default disabled)
  -G, --gid int           set the group of the files in the filesystem (default disabled)
  -f, --foreground        do not demonize
  -d, --debug             activate debug output (implies --foreground)
  -h, --help              display this help and exit


$ cat ~/.pulsar-fuse.yaml
---
pulsar-fuse:
  topic1:
    url: "pulsar://localhost:6650"
    type: persistent # or non-persistent. default value is persistent
    tenant: public # default value is public
    namespace: default # default value is default
    topic: my-topic
    compression: lz4 # zlib,zstd,snappy. default value is none
    log_name: abc.log
  topic2:
    url: yyy
    log_name: zxy.log
// EOF

# Write a log
$ echo '{"msg": "error"}' > /tmp/mountpoint/abc.log

# Produce a log message('{"msg": "error"}') to topic1

$ ls -Fla /tmp/mountpoint/abc.log
-rw-r--r-- 0 root root 8 Jul  7 07:07 abc.log

$ cat /tmp/mountpoint/abc.log
{{EMPTY_CONTENT}}

$ rm /tmp/mountpoint/abc.log
rm: cannot access '/tmp/mountpoint/abc.log': Permission denied

$ ls -Fla /tmp/mountpoint/abc.log

```

## Developed by VSCode-Remote-Development-Containers

```sh
# Open
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

$ sudo cat /etc/fuse.conf
# /etc/fuse.conf - Configuration file for Filesystem in Userspace (FUSE)

# Set the maximum number of FUSE mounts allowed to non-root users.
# The default is 1000.
#mount_max = 1000

# Allow non-root users to specify the allow_other or allow_root mount options.
#user_allow_other

$ sudo sed -i -e 's/#user_allow_other/user_allow_other/g' /etc/fuse.conf

$ grep user_allow_other /etc/fuse.conf
user_allow_other

$ mkdir /tmp/mountpoint
$ sudo ./hello /tmp/mountpoint &
sudo ls -Fla /tmp/mountpoint &
sudo cat /tmp/mountpoint/file.txt

$ sudo ./hello /tmp/mountpoint &
sudo echo google >> /tmp/mountpoint/file.txt
bash: /tmp/mountpoint/file.txt: Permission denied

# 上記コマンド実行時に以下のエラーが出る場合は、fusermountコマンドを実行
#/bin/fusermount: failed to access mountpoint /tmp/mountpoint: Transport endpoint is not connected
$ sudo fusermount -u /tmp/mountpoint
```

## Developed by Mac

1. Virtual Box
2. Vagrant

```sh
$ brew cask install virtualbox
$ brew cask install vagrant
$ vagrant -v
$ vagrant plugin install vagrant-disksize vagrant-hostsupdater vagrant-mutagen
```

## Docker for Mac

### Build

```sh
$ docker run --rm -it -v $(pwd):/go/src/github.com/transnano/go-fuse-sample/ -w /go/src/github.com/transnano/go-fuse-sample golang:1.14.4 bash

docker run --rm -it -v $(pwd):/go/src/github.com/transnano/go-fuse-sample/ -w /go/src/github.com/transnano/go-fuse-sample golang:1.14.4 go build -o hello main.go
```

### Check

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
