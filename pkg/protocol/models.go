package protocol

type RR struct{
	Header
	RData []byte
}

type Header struct{
	Name string
	Type uint16
	Class uint16
	TTL uint32
	RDLength uint16
}