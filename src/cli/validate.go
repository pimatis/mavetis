package cli

import (
	"errors"

	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/model"
)

func validateReview(spec model.Review) error {
	if len(spec.Files) != 0 && spec.Staged {
		return errors.New("@file targets cannot be combined with --staged")
	}
	if len(spec.Files) != 0 && spec.Base != "" {
		return errors.New("@file targets cannot be combined with --base")
	}
	if len(spec.Files) != 0 && spec.Head != "" {
		return errors.New("@file targets cannot be combined with --head")
	}
	if spec.WithSuggested && len(spec.Files) == 0 {
		return errors.New("review option --with-suggested requires file targets")
	}
	if spec.StdinTargets && len(spec.Files) == 0 {
		return errors.New("review option --stdin-targets requires newline-separated paths on stdin")
	}
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
