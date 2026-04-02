package cli

import (
	"errors"

	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/model"
)

func validateReview(spec model.Review) error {
	if spec.Staged && spec.Base != "" {
		return errors.New("review options --staged and --base cannot be combined")
	}
	if spec.Staged && spec.Head != "" {
		return errors.New("review option --head requires --base")
	}
	if spec.Head != "" && spec.Base == "" {
		return errors.New("review option --head requires --base")
	}
	if err := config.ValidateOutput(spec.Format); err != nil {
		return err
	}
	if err := config.ValidateSeverity(spec.Severity, "severity"); err != nil {
		return err
	}
	if err := config.ValidateSeverity(spec.FailOn, "fail-on"); err != nil {
		return err
	}
	if err := config.ValidateProfile(spec.Profile); err != nil {
		return err
	}
	return nil
}
