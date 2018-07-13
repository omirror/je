package je

import (
	"fmt"
	"strconv"
	"strings"
)

type URI struct {
	Type string
	Path string
}

func (u *URI) String() string {
	return fmt.Sprintf("%s://%s", u.Type, u.Path)
}

func ParseURI(uri string) (*URI, error) {
	parts := strings.Split(uri, "://")
	if len(parts) == 2 {
		return &URI{Type: strings.ToLower(parts[0]), Path: parts[1]}, nil
	}
	return nil, fmt.Errorf("invalid uri: %s", uri)
}

// SafeParseInt ...
func SafeParseInt(s string, d int) int {
	n, e := strconv.Atoi(s)
	if e != nil {
		return d
	}
	return n
}

// SafeParseUint64 ...
func SafeParseUint64(s string, d uint64) uint64 {
	n, e := strconv.ParseUint(s, 10, 64)
	if e != nil {
		return d
	}
	return n
}
