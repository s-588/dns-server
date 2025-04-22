package protocol

type flagOptions struct {
	// See RFC 1035 25-27 pages for details
	// 0... .... .... .... = QR: Message is a query,
	// .000 0... .... .... = OPCODE: Standard query (0),
	// .... .0.. .... .... = AA:  Authoritative Answer
	// .... ..0. .... .... = TC: Message is not truncated,
	// .... ...1 .... .... = RD: Do query recursively,
	// .... .... 1... .... = RA: Recursion AvaRecursion Availableilable
	// .... .... .0.. .... = Z: reserved (0),
	// .... .... ..00 00.. = RCODE: No error condition

	// Is this response? 0 - query, 1 - response
	query bool

	// What kind of query is this. 0 - standard, 1 - inversed, 2 - server status, 3-15 reserved.
	opcode byte

	// Is query comming from authoritative resource
	authoritative bool

	// Is query was truncated due to length greater than that permitted on the transmission channel.
	truncated bool

	// TODO: continue creating of this file
}
