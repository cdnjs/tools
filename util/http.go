package util

// GetProtocol gets the protocol, either http or https.
func GetProtocol() string {
	if HasHTTPProxy() {
		return "http"
	}
	return "https"
}
