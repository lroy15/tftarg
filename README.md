### Tftarg
 
A simple wrapper around terraform that grabs resources from git diff and lets you select the ones you want to plan/apply.

Download the binary from the release page, or clone the repo and run:

- `go mod tidy`
- `go build` and add the path to the binary to your path, for example add this to your .zshrc file: export PATH=$PATH:"path/to/tftarg/binary"
 
OR just run `go install` - This will build and place the binary in your Go binary path (default is $GOPATH/bin or $HOME/go/bin in modern Go setups). You can run the program from any directory once it's installed.


