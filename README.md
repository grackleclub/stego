# cryptogif

Method for encoding very small amounts of data, with incredible ineffiency, into a gif's color palette assignment.

## why?

Just because.

## how?

Each frame of a gif has a 256 bit color space. This tool reserves the darkest 16 colors (as determined by the sum of red, green, and blue values) for encoding data, arbitrarily reassigning any existing use of reserved colors to the 17th darkest palette index. This can create artifacts, but the hope is that darker colors are less noticible. The data is then encoded in what is essentially hexadecimal, as determined by the darkness index of each frame's palette.

> [!WARNING]
> This obfuscation tool is **not** a substitute for encryption.

### write data to a gif
```go
originalGif, _ := cryptogif.Read("./path/to/file.gif")
myData := []byte("here's some arbitrary data!")
modifiedGif, _ := cryptogif.Inject(originalGif, myData)

_ = cryptogif.Write(modifiedGif, "./path/to/new-file.gif")
```

### read data from a gif
```go
gif, _ := cryptogif.Read("./path/to/file.gif")
data, _ := cryptogif.Extract(gif)
```
