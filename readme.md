pwdgo changes Go toolchains as you change directories.

There are lots of ways to accomplish this task. Many involve dropping ".env" files all over the place.

This program is another approach: I infer the Go toolchain version to use from the go.mod file. (Then you can't accidentally introduce Go 1.16 features without bumping the Go version in go.mod.)

I wrote it for myself, to match my own peculiar setup.

pwdgo emits a new suggested path based on the current directory.

I use it with zsh's `chpwd` function that runs every time you change directories. Sample usage:

```
function chpwd {
    NEWPATH=$(/Users/josh/bin/pwdgo -go 1.15:/Users/josh/go/1.15/bin -go 1.16:/Users/josh/go/1.16/bin -go ts:/Users/josh/go/ts/bin -go tip:/Users/josh/go/tip/bin -dir /Users/josh/go/tip:tip -dir /Users/josh/go/1.16:1.16 -path tailscale.io:ts -path tailscale.com:ts -default 1.16)
    if [ "$NEWPATH" != "" ];then
        export PATH="${NEWPATH}"
    fi
}
```

This defines a bunch of Go toolchain locations based on Go versions. It defines some overrides for particular Go modules and particular directories on disk. It provides a default toolchain if there is no go.mod (if the go.mod version is unrecognized) and none of the overrides match.
