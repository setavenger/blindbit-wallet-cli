package utils

import (
	"fmt"
)

// todo: can be expanded to handle full host names with dns etc.
func ConstructBaseUrl(host string, port int) string {
	return fmt.Sprintf("http://%s:%d", host, port)
}
