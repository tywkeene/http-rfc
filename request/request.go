package request

import (
	"net/url"
)

type Request struct {
	Method       string
	URL          *url.URL
	ProtoVersion string
	ProtoMajor   int
	ProtoMinor   int
}
