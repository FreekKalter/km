package km

import "fmt"
import "regexp"

// Response a custum error response that gives some more ditails on the error
type Response struct {
	Code  int
	Regex *regexp.Regexp
	Extra string
}

// String impletent the Stringer interface for easy printing
func (r Response) String() string {
	return fmt.Sprintf("%s : %s", r.Regex.String(), r.Extra)
}

// Error implements the error interface
func (r Response) Error() string {
	return r.Regex.String()
}

// newResponse creates a new Response object
func newResponse(regex string, code int) Response {
	r := Response{Code: code}
	r.Regex = regexp.MustCompile(regex)
	return r
}

var (
	// NotFound standard 404 not found
	NotFound = newResponse("^404 page not found\n$", 404)
	// Ok 200 ok
	Ok = newResponse("ok\n", 200)
	// UnknownField 400 an unknown field encountered in supplied data
	UnknownField = newResponse("invalid fieldname\n", 400)
	// NotParsable 400 could not parse request
	NotParsable = newResponse("could not parse request\n", 400)
	// InvalidDate coudl not parse the date provided
	InvalidDate = newResponse("invalid date\n", 400)
	// InvalidURL 400 invalid url, correct structure, but invalid
	InvalidURL = newResponse("invalid url", 400)
	// DbError error connecting to database
	DbError = newResponse("database eror", 500)
)

// CustomResponse takes a error and adds extra fields to convert it to a custom Response object
func CustomResponse(r Response, err error) Response {
	ret := r
	ret.Extra = err.Error()
	return ret
}
