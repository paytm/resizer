# Resizer 

Resizer is an http image resizer. It uses magickwand to resize images on the fly.
Source images are obtained from an upstream (file/http). They can optionally be cached to a downstram (file/s3).

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
