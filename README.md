# Resizer 

Resizer is an http image resizer. It uses magickwand to resize images on the fly.
Source images are obtained from an upstream (file/http). They can optionally be cached to a downstream (file/s3).

## Getting Started

Set your GOPATH, then install using

~~~
go get github.com/qzaidi/resizer
~~~

This will create $GOPATH/bin/resizer, which is an http server.

Before running resizer binary, make sure to create a config file (in CWD or in /etc). There's a sample config in cfg/

~~~
[Server]
Port = 4000 # Port to listen on

[Upstream]
URI = http://catalogadmin.paytm.com # Server to download source images from

[Downstream]
URI = s3://ABCDEFGHIJ:yoursecrets3keygoeshere@assets.paytm.com # S3 server to save resized images to
~~~

## Pre-commit hook

The following hook should be copied to .git/hooks/pre-commit to enable versioning.

~~~
#!/bin/sh
 
#picked from  http://alimoeenysbrain.blogspot.com/2013/10/automatic-versioning-with-git-commit.html
VERBASE=$(git rev-parse --verify HEAD | cut -c 1-7)
echo $VERBASE
NUMVER=$(awk '{printf("%s", $0); next}' version.go | sed 's/.*Resizer\.//' | sed 's/\..*//')
echo "old version: $NUMVER"
NEWVER=$(expr $NUMVER + 1)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "new version: Resizer.$NEWVER.$VERBASE"
BODY="package main\n\nfunc ResizerVersion() string {\n\treturn \"Resizer.$NEWVER.$BRANCH.$VERBASE\"\n}\n"
echo $BODY > version.go
git add version.go
~~~
