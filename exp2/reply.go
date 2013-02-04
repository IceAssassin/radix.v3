package redis

import (
	"errors"
	"strconv"
)

//* Reply

/*
ReplyType describes type of a reply.

Possible values are:

StatusReply --  status reply
ErrorReply -- error reply
IntegerReply -- integer reply
NilReply -- nil reply
BulkReply -- bulk reply
MultiReply -- multi bulk reply
*/
type ReplyType uint8

const (
	StatusReply ReplyType = iota
	ErrorReply
	IntegerReply
	NilReply
	BulkReply
	MultiReply
)

// Reply holds a Redis reply.
type Reply struct {
	Type  ReplyType // Reply type
	Elems []*Reply  // Sub-replies
	Err   error    // Reply error
	str   string
	int   int64
}

// Str returns the reply value as a string or
// an error, if the reply type is not StatusReply or BulkReply.
func (r *Reply) Str() (string, error) {
	if r.Type == ErrorReply {
		return "", r.Err
	}
	if !(r.Type == StatusReply || r.Type == BulkReply) {
		return "", errors.New("string value is not available for this reply type")
	}

	return r.str, nil
}

// Bytes is a convenience method for calling Reply.Str() and converting it to []byte.
func (r *Reply) Bytes() ([]byte, error) {
	s, err := r.Str()
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

// Int64 returns the reply value as a int64 or an error,
// if the reply type is not IntegerReply or the reply type
// BulkReply could not be parsed to an int64.
func (r *Reply) Int64() (int64, error) {
	if r.Type == ErrorReply {
		return 0, r.Err
	}
	if r.Type != IntegerReply {
		s, err := r.Str()
		if err == nil {
			i64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return 0, errors.New("failed to parse integer value from string value")
			} else {
				return i64, nil
			}
		}

		return 0, errors.New("integer value is not available for this reply type")
	}

	return r.int, nil
}

// Int is a convenience method for calling Reply.Int64() and converting it to int.
func (r *Reply) Int() (int, error) {
	i64, err := r.Int64()
	if err != nil {
		return 0, err
	}

	return int(i64), nil
}

// Bool returns false, if the reply value equals to 0 or "0", otherwise true; or
// an error, if the reply type is not IntegerReply or BulkReply.
func (r *Reply) Bool() (bool, error) {
	if r.Type == ErrorReply {
		return false, r.Err
	}
	i, err := r.Int()
	if err == nil {
		if i == 0 {
			return false, nil
		}

		return true, nil
	}

	s, err := r.Str()
	if err == nil {
		if s == "0" {
			return false, nil
		}

		return true, nil
	}

	return false, errors.New("boolean value is not available for this reply type")
}

// List returns a multi bulk reply as a slice of strings or an error.
// The reply type must be MultiReply and its elements' types must all be either BulkReply or NilReply.
// Nil elements are returned as empty strings.
// Useful for list commands.
func (r *Reply) List() ([]string, error) {
	if r.Type == ErrorReply {
		return nil, r.Err
	}
	if r.Type != MultiReply {
		return nil, errors.New("reply type is not MultiReply")
	}

	strings := make([]string, len(r.Elems))
	for i, v := range r.Elems {
		if v.Type == BulkReply {
			strings[i] = v.str
		} else if v.Type == NilReply {
			strings[i] = ""
		} else {
			return nil, errors.New("element type is not BulkReply or NilReply")
		}
	}

	return strings, nil
}

// Hash returns a multi bulk reply as a map[string]string or an error.
// The reply type must be MultiReply, 
// it must have an even number of elements,
// they must be in a "key value key value..." order and
// values must all be either BulkReply or NilReply.
// Nil values are returned as empty strings.
// Useful for hash commands.
func (r *Reply) Hash() (map[string]string, error) {
	if r.Type == ErrorReply {
		return nil, r.Err
	}
	rmap := map[string]string{}

	if r.Type != MultiReply {
		return nil, errors.New("reply type is not MultiReply")
	}

	if len(r.Elems)%2 != 0 {
		return nil, errors.New("reply has odd number of elements")
	}

	for i := 0; i < len(r.Elems)/2; i++ {
		var val string

		key, err := r.Elems[i*2].Str()
		if err != nil {
			return nil, errors.New("key element has no string reply")
		}

		v := r.Elems[i*2+1]
		if v.Type == BulkReply {
			val = v.str
			rmap[key] = val
		} else if v.Type == NilReply {
		} else {
			return nil, errors.New("value element type is not BulkReply or NilReply")
		}
	}

	return rmap, nil
}

// String returns a string representation of the reply and its sub-replies.
// This method is for debugging.
// Use method Reply.Str() for reading string reply.
func (r *Reply) String() string {
	switch r.Type {
	case ErrorReply:
		return r.Err.Error()
	case StatusReply:
		fallthrough
	case BulkReply:
		return r.str
	case IntegerReply:
		return strconv.FormatInt(r.int, 10)
	case NilReply:
		return "<nil>"
	case MultiReply:
		s := "[ "
		for _, e := range r.Elems {
			s = s + e.String() + " "
		}
		return s + "]"
	}

	// This should never execute
	return ""
}