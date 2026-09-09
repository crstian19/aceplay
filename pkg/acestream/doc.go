// Package acestream talks to acestream-engine over HTTP and parses
// acestream:// URLs.
//
// It is the reusable half of aceplay and carries no dependency on the CLI:
// URL parsing ([ParseURL]) and the engine HTTP API ([Client]) are all it does.
// Managing the local engine process (finding the binary, spawning and killing
// it) lives in aceplay's internal packages, because that part is not reusable.
//
// A zero-config client talks to a local engine on the default port:
//
//	client := acestream.NewClient()
//	url, err := client.WaitForStream(ctx, contentID)
package acestream
