package utils

import (
	"io"

	"github.com/json-iterator/go"
	"github.com/mailru/easyjson"
)

//var json = jsoniter.ConfigCompatibleWithStandardLibrary

func JsonEncoder(w io.Writer, v interface{}) error {
	if m, ok := v.(easyjson.Marshaler); ok {
		_, err := easyjson.MarshalToWriter(m, w)
		return err
	}
	return jsoniter.NewEncoder(w).Encode(v)
	//return json.NewEncoder(w).Encode(v)
}

func JsonUnmarshal(d []byte, v interface{}) error {
	if m, ok := v.(easyjson.Unmarshaler); ok {
		return easyjson.Unmarshal(d, m)
	}
	return jsoniter.Unmarshal(d, v)
}

func JsonMarshal(v interface{}) ([]byte, error) {
	if m, ok := v.(easyjson.Marshaler); ok {
		return easyjson.Marshal(m)
	}
	return jsoniter.Marshal(v)
}

//func UnmarshalJsonUseNumber(data []byte, v interface{}) error {
//	dec := json.NewDecoder(bytes.NewReader(data))
//	dec.UseNumber()
//	return dec.Decode(v)
//}
