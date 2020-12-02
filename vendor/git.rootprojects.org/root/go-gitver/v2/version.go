package main

// use recently generated version info as a fallback
// for when git isn't present (i.e. go run <url>)
func init() {
	commit = "37c1fd4b5694fd62c9f0d6ad1df47d938accbeec"
	version = "2.0.0-pre1-dirty"
	date = "2020-10-10T16:05:59-06:00"
}
