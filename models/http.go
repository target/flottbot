package models

// HTTPResponse base HTTP response data structure
type HTTPResponse struct {
	Status int
	Raw    string
	Data   interface{}
}
