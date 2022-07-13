package resp

import "encoding/json"

type ResponseCode string

const (
	ResponseCodeSuccess ResponseCode = "Success"
)

// TODO: Other values?

func (rc *ResponseCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*rc = ResponseCode(str)
	return nil
}
