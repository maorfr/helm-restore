# Helm Restore Plugin

This is a Helm plugin which restores the last deployed release to its original state

## Usage

restore last deployed release to original state

```
$ helm restore [flags] RELEASE_NAME
```

### Flags:

```
  -l, --label string              label to select tiller resources by (default "OWNER=TILLER,STATUS=DEPLOYED")
      --tiller-namespace string   namespace of Tiller (default "kube-system")   
```

## Install

```
$ helm plugin install https://github.com/maorfr/helm-restore
```

The above will fetch the latest binary release of `helm restore` and install it.

### Developer (From Source) Install

If you would like to handle the build yourself, instead of fetching a binary,
this is how recommend doing it.

First, set up your environment:

- You need to have [Go](http://golang.org) installed. Make sure to set `$GOPATH`
- If you don't have [Dep](https://github.com/golang/dep) installed, this will install it into
  `$GOPATH/bin` for you.

Clone this repo into your `$GOPATH`. You can use `go get -d github.com/maorfr/helm-restore`
for that.

```
$ cd $GOPATH/src/github.com/maorfr/helm-restore
$ make bootstrap build
$ SKIP_BIN_INSTALL=1 helm plugin install $GOPATH/src/github.com/maorfr/helm-restore
```

That last command will skip fetching the binary install and use the one you
built.
