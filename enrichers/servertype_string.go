// Code generated by "stringer -type=ServerType"; DO NOT EDIT.

package enrichers

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Unknown-0]
	_ = x[ELK-1]
	_ = x[FTP-2]
	_ = x[SSH-3]
	_ = x[SQL-4]
	_ = x[HTTP-5]
}

const _ServerType_name = "UnknownELKFTPSSHSQLHTTP"

var _ServerType_index = [...]uint8{0, 7, 10, 13, 16, 19, 23}

func (i ServerType) String() string {
	if i < 0 || i >= ServerType(len(_ServerType_index)-1) {
		return "ServerType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ServerType_name[_ServerType_index[i]:_ServerType_index[i+1]]
}