# PegNet API Auditor

## Building

You'll need `go 1.13` installed but otherwise there are no special requirements. Just run `go build .` in the directory.

## Setup & Running

The app looks for [config.ini](https://github.com/WhoSoup/pnmc-auditor/blob/master/config.ini) in the working directory or the PATH. The configuration options are:

* `paying`: This is an EC address key that is used to pay for the entries written to a chain
* `signing`: This is an FA address key that is used to crytographically sign the entries. The public key will be published along with the signature
* `chain`: This is the chain to write entries to. It has to be created separately
* `factomd`: This points to the open node. If you want to use a different node, change this
* `interval`: This is the time in **seconds** that it will write entries. A 60-second interval will cost 1440 EC a day.

To run, just start the executable.