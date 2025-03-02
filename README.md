# stego

[![Go Test](https://github.com/grackleclub/stego/actions/workflows/go.yml/badge.svg)](https://github.com/grackleclub/stego/actions/workflows/go.yml)

Method for encoding very small amounts of data, with incredible ineffiency, into a gif's color palette assignment.

## why?

Just because.

## how?

Each frame of a gif has a 256 bit color space. This tool reserves the darkest 16 colors (as determined by the sum of red, green, and blue values) for encoding data, arbitrarily reassigning any existing use of reserved colors to the 17th darkest palette index. This can create artifacts, but the hope is that darker colors are less noticible. The data is then encoded in what is essentially hexadecimal, as determined by the darkness index of each frame's palette.

> [!WARNING]
> This obfuscation tool is **not** a substitute for encryption.

## examples

original | embedded with "Hello, world!"
--- | ---
![original](./img/originals/earth.gif) | ![modified](./img/output/earth_output.gif)

### write data to a gif
```go
// read the file
originalGif, _ := stego.Read("./path/to/file.gif")
// define some data as a slice of bytes
myData := []byte("here's some arbitrary data!")
// inject the data into a modified gif
modifiedGif, _ := stego.Inject(originalGif, myData)
// write the new gif to file
_ = stego.Write(modifiedGif, "./path/to/new-file.gif")
```

### read data from a gif
```go
// read the file
gif, _ := stego.Read("./path/to/file.gif")
// extract the data as a slice of bytes
data, _ := stego.Extract(gif)
fmt.Printf("extracted some data: %s", string(data))
```
