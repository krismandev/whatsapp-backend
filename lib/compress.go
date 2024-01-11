package lib

import (
	"bytes"
	"compress/zlib"
	"strings"
)

func Compress(s string) string {
	buff := bytes.Buffer{}
	compressEnc := zlib.NewWriter(&buff)
	compressEnc.Write([]byte(s))
	compressEnc.Close()
	//return base64.StdEncoding.EncodeToString(buff.Bytes())
	return string(buff.Bytes())
}

// func Compress(s string) string {
// 	buff := bytes.Buffer{}
// 	compressEnc := gzip.NewWriter(&buff)
// 	compressEnc.Write([]byte(s))
// 	compressEnc.Close()
// 	//return base64.StdEncoding.EncodeToString(buff.Bytes())
// 	return string(buff.Bytes())
// }

func Decompress(s string) (string, error) {
	// b, err := base64.StdEncoding.DecodeString(s)
	// if err != nil {
	// 	return "", err
	// }
	//decompressor, err := zlib.NewReader(bytes.NewReader(b))

	decompressor, err := zlib.NewReader(strings.NewReader(s))
	if err != nil {
		return "", err
	}
	var decompressedBuff bytes.Buffer
	decompressedBuff.ReadFrom(decompressor)
	return decompressedBuff.String(), nil
}

func CompressBytes(s []byte) []byte {
	buff := bytes.Buffer{}
	compressEnc := zlib.NewWriter(&buff)
	compressEnc.Write(s)
	compressEnc.Close()

	//return []byte(base64.StdEncoding.EncodeToString(buff.Bytes()))
	return buff.Bytes()
}

func DecompressBytes(s string) ([]byte, error) {
	// b, err := base64.StdEncoding.DecodeString(s)
	// if err != nil {
	// 	return nil, err
	// }
	// decompressor, err := zlib.NewReader(bytes.NewReader(b))
	decompressor, err := zlib.NewReader(strings.NewReader(s))
	if err != nil {
		return nil, err
	}

	var decompressedBuff bytes.Buffer
	decompressedBuff.ReadFrom(decompressor)
	return decompressedBuff.Bytes(), nil
}
