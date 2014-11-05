package km

import "fmt"
import "regexp"

type Response struct {
	Code  int
	Regex *regexp.Regexp
	Extra string
}

func (r Response) String() string {
	return fmt.Sprintf("%s : %s", r.Regex.String(), r.Extra)
}

func (r Response) Error() string {
	return r.Regex.String()
}

func newResponse(regex string, code int) Response {
	r := Response{Code: code}
	r.Regex = regexp.MustCompile(regex)
	return r
}

var NotFound = newResponse("^404 page not found\n$", 404)
var Ok Response = newResponse("ok\n", 200)
var OkId Response = newResponse("[0-9]+", 200)
var UnknownField Response = newResponse("invalid fieldname\n", 400)
var NotParsable Response = newResponse("could not parse request\n", 400)
var InvalidDate Response = newResponse("invalid date\n", 400)
var InvalidUrl Response = newResponse("invalid url", 400)
var DbError Response = newResponse("database eror", 500)

func CustomResponse(r Response, err error) Response {
	ret := r
	ret.Extra = err.Error()
	return ret
}
