package output

import (
	"encoding/json"

	"github.com/Pimatis/mavetis/src/model"
)

func JSON(report model.Report) (string, error) {
	buffer, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}
