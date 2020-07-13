package util

// Gets the protocol, either http or https.
func GetProtocol() string {
	if HasHTTPProxy() {
		return "http"
	}
	return "https"
}
