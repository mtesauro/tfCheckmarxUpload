NOTE: This is an example of what the setup of tfCheckmarx-uploader could look like.  The binary in this directory has been compiled for 64-bit Linux installs and will only work in that situation.  A binary for your OS + Architecture will need to be compiled from the Go source if your installation target differs.

For example, to compile for Windows 64-bit, you would issue the command:

> GOOS=windows GOARCH=amd64 go build

in the root directory of this repository to compile tfCheckmarx-uploader.go to tfCheckmarx-uploader.exe.


