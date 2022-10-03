package handler

// TODO: copy this documentation to http-wasm-abi, then cite the links.

const (
	// HostModule is the WebAssembly module name of the functions the host
	// exports.
	HostModule = "http-handler"

	// FuncLog logs a message to the host's logs.
	//
	// # Parameters
	//
	// All parameters are of type i32. They contain the log message.
	//
	//   - message: memory offset of the UTF-8 encoded message.
	//   - message_len: possibly zero length of the message in bytes.
	//
	// # Result
	//
	// There is no result from this function. A host who fails to log the
	// message will trap (aka panic, "unreachable" instruction).
	//
	// # Example
	//
	// For example, if parameters are message=8, message_len=2, this function
	// would log the message "error".
	//
	//	              message_len
	//	           +--------------+
	//	           |              |
	//	[]byte{?, 'e', 'r', 'o', 'r', ?}
	//   message --^
	FuncLog = "log"

	// FuncHandle is what the guest exports to handle an HTTP server request.
	//
	// # Parameters
	//
	// There are no parameters
	//
	// # Result
	//
	// There is no result from this function. A guest who fails to handle the
	// request will trap (aka panic, "unreachable" instruction).
	FuncHandle = "handle"

	// FuncGetRequestHeader writes a header value to memory if it exists and
	// isn't larger than the buffer size limit. The result is `1<<32|value_len`
	// or zero if the header doesn't exist. `value_len` is the actual value
	// length in bytes.
	//
	// # Use cases
	//
	// This signature supports the most common case of retrieving a header
	// value by name. However, there are some subtle use cases possible due to
	// the signature design, particularly helpful for WebAssembly performance:
	//
	//   - re-using a buffer for multiple header reads (`buf`).
	//   - growing a buffer only when needed (retry with larger `buf_limit`).
	//   - avoiding copying invalidly large header values (`buf_limit`).
	//   - determining if a header exists without copying it (`buf_limit=0`).
	//
	// # Parameters
	//
	// All parameters are of type i32. They contain the UTF-8 header name and
	// a buffer to write the value.
	//
	//   - name: memory offset to read the header name.
	//   - name_len: length of the header name in bytes.
	//   - buf: memory offset to write the header value, if exists and not
	//     larger than `buf_limit` bytes.
	//   - buf_limit: possibly zero maximum length in bytes to write. If the
	//     result `value_len` is larger, nothing is written to memory.
	//
	// Note: Hosts will compare the name case insensitively to adhere to HTTP
	// semantics.
	//
	// # Result
	//
	// Both results are of type i32. A host who fails to read the request
	// header will trap (aka panic, "unreachable" instruction).
	//
	//   - exists: zero if the header does not exist and one if it does.
	//   - value_len: possibly zero length in bytes of the header value.
	//
	// To retain compatability with WebAssembly 1.0, both results are packed
	// into a single i64 result.
	//
	// If the result is zero, there is no value. Otherwise, the lower 32-bits
	// are `value_len`. For example, in WebAssembly `i32.wrap_i64` unpacks
	// `value_len` as would casting in most languages (ex `uint32(result)` in
	// Go).
	//
	// # Example
	//
	// For example, if parameters name=1 and name_len=4, this function would
	// read the header name "ETag".
	//
	//	               name_len
	//	           +--------------+
	//	           |              |
	//	[]byte{?, 'E', 'T', 'a', 'g', ?}
	//	    name --^
	//
	// If there was no `ETag` header, the result would be i64(0) and the user
	// doesn't need to read memory.
	//
	// If it exists and is "01234567", then `value_len=8`, so the result is
	// i64(1<<32 | 8) or i64(4294967304). If the `buf_limit` parameter was 7,
	// nothing would be written to memory. The caller would decide whether to
	// retry the request with a higher limit.
	//
	// If parameters buf=16 and buf_limit=128, the value would be written to
	// memory like so:
	//
	//	                              value_len
	//	                +----------------------------------+
	//	                |                                  |
	//	[]byte{ 0..15, '0', '1', '2', '3', '4', '5', '6', '7', ?, .. }
	//	          buf --^
	FuncGetRequestHeader = "get_request_header"

	// FuncGetPath writes the request path value to memory, if it isn't larger
	// than the buffer size limit. The result is the actual path length in
	// bytes.
	//
	// # Use cases
	//
	// This signature supports the most common case of retrieving the request
	// path. However, there are some subtle use cases possible due to
	// the signature design, particularly helpful for WebAssembly performance:
	//
	//   - re-using a buffer for multiple path reads (`buf`).
	//   - growing a buffer only when needed (retry with larger `buf_limit`).
	//   - avoiding copying invalidly large paths (`buf_limit`).
	//
	// # Parameters
	//
	// All parameters are of type i32.
	//
	//   - buf: memory offset to write the path, if not larger than `buf_limit`
	//     bytes.
	//   - buf_limit: possibly zero maximum length in bytes to write. If the
	//     result (`path_len`) is larger, nothing is written to memory.
	//
	// Note: The path does not include query parameters.
	//
	// # Result
	//
	// The result, `path_len`, is the i32 path length in bytes, even if larger
	// than `buf_limit`. A host who fails to read the request path will trap
	// (aka panic, "unreachable" instruction).
	//
	// # Example
	//
	// For example, if parameters buf=16 and buf_limit=128, and the request
	// line was "GET /foo?bar", "/foo" would be written to memory like so:
	//
	//	                    path_len
	//	                +--------------+
	//	                |              |
	//	[]byte{ 0..15, '/', 'f', 'o', 'o', ?, .. }
	//	          buf --^
	FuncGetPath = "get_path"

	// FuncSetPath overwrites the request path with one read from memory.
	//
	// # Parameters
	//
	// All parameters are of type i32. They contain the UTF-8 path.
	//
	//   - path: memory offset to set the path.
	//   - path_len: possibly zero length of the path in bytes.
	//
	// # Result
	//
	// There is no result from this function. A host who fails to set the path
	// will trap (aka panic, "unreachable" instruction).
	//
	// # Example
	//
	// For example, if parameters are path=8, path_len=2, this function would
	// set the path to "/a".
	//
	//	          path_len
	//	           +----+
	//	           |    |
	//	[]byte{?, '/', 'a', ?}
	//	    path --^
	FuncSetPath = "set_path"

	// FuncNext is an alternative to FuncSendResponse that dispatches control
	// to the next handler on the host.
	//
	// # Parameters
	//
	// There are no parameters
	//
	// # Result
	//
	// There is no result from this function. A host who fails to dispatch to
	// the next handler will trap (aka panic, "unreachable" instruction).
	FuncNext = "next"

	// FuncSetResponseHeader sets a response header from a name and value read
	// from memory.
	//
	// # Parameters
	//
	// All parameters are of type i32. They contain the UTF-8 header name and
	// value of the response header.
	//
	//   - name: memory offset to set the header name.
	//   - name_len: length of the header name in bytes.
	//   - value: memory offset to set the header value.
	//   - value_len: possibly zero length of the header value in bytes.
	//
	// # Result
	//
	// There is no result from this function. A host who fails to set a value
	// will trap (aka panic, "unreachable" instruction).
	//
	// # Example
	//
	// For example, if parameters are name=1, name_len=4, value=8, value_len=1,
	// this function would set the response header "ETag: 1".
	//
	//	               name_len             value_len
	//	           +--------------+             +
	//	           |              |             |
	//	[]byte{?, 'E', 'T', 'a', 'g', ?, ?, ?, '1', ?}
	//	    name --^                            ^
	//	                                value --+
	FuncSetResponseHeader = "set_response_header"

	// FuncSendResponse is an alternative to FuncNext that sends the HTTP
	// response with a given status code and optional body.
	//
	// # Parameters
	//
	// All parameters are of type i32. These describe the status code and the
	// optional body to send.
	//
	//   - status_code: HTTP status code. Ex. 200
	//   - body: memory offset of the response body.
	//   - body_len: possibly zero length of the body in bytes.
	//
	// Note: The "Content-Length" header is set to `body_len` when non-zero.
	// If you need to set "Content-Length: 0", call FuncSetResponseHeader prior
	// to this.
	//
	// # Result
	//
	// There is no result from this function. A host who fails to send the body
	// will trap (aka panic, "unreachable" instruction).
	//
	// # Example
	//
	// For example, if parameters are status_code=401, body=1, body_len=0,
	// this function would send the HTTP status code 401 with no body or
	// "Content-Length" header.
	FuncSendResponse = "send_response"
)
