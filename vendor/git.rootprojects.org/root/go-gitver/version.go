package main

// use recently generated version info as a fallback
// for when git isn't present (i.e. go run <url>)
func init() {
	GitRev = "0b8c2d86df4bfe32ff4534eec84cd56909c398e9"
	GitVersion = "v1.1.1"
	GitTimestamp = "2019-06-21T00:18:13-06:00"
}
