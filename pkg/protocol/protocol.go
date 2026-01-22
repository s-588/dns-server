package protocol

import "encoding/binary"

func ReadRR(data []byte) (RR, error) {

}

func WriteRR(rr RR) ([]byte, error) {
}

func readHeader(data []byte) (Header, error) {
	h := Header{
	}
}

func writeHeader(header Header) ([]byte, error) {
	
}

func readUint16(msg []byte) uint16 {
	return binary.BigEndian.Uint16(msg)
}

func writeUint16(msg []byte, value uint16) []byte {
	return binary.BigEndian.AppendUint16(msg, value)
}