package atfuck

//dir to store some cached files for atfuck, like ak, sk
var QShellRootPath string

const (
	BLOCK_BITS = 22 // Indicate that the blocksize is 4M
	BLOCK_SIZE = 1 << BLOCK_BITS
)

const (
	STATUS_OK = iota
	//process error
	STATUS_ERROR
	//local error
	STATUS_HALT
)
