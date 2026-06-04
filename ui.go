package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

// selectPipelineDefinition asks the user to pick one pipeline definition.
func selectPipelineDefinition(defs []PipelineDefinition) (PipelineDefinition, error) {
	if len(defs) == 0 {
		return PipelineDefinition{}, errors.New("no pipeline definitions found")
	}
	opts := make([]huh.Option[string], len(defs))
	byID := make(map[string]PipelineDefinition, len(defs))
	for i, d := range defs {
		label := d.Name
		if d.Description != "" {
			label = fmt.Sprintf("%s — %s", d.Name, d.Description)
		}
		opts[i] = huh.NewOption(label, d.ID)
		byID[d.ID] = d
	}

	var id string
	err := huh.NewSelect[string]().
		Title("Pipeline definition").
		Description("Select which pipeline does the publishing of your package.").
		Options(opts...).
		Value(&id).
		Run()
	if err != nil {
		return PipelineDefinition{}, err
	}
	return byID[id], nil
}

// selectContexts asks the user to pick zero or more contexts (filterable).
func selectContexts(ctxs []Context) ([]string, error) {
	if len(ctxs) == 0 {
		return nil, nil
	}
	opts := make([]huh.Option[string], len(ctxs))
	for i, c := range ctxs {
		opts[i] = huh.NewOption(c.Name, c.ID)
	}

	var ids []string
	err := huh.NewMultiSelect[string]().
		Title("Contexts").
		Description("Press / to filter, space to toggle, enter to confirm. Selecting none is OK.").
		Options(opts...).
		Filterable(true).
		Value(&ids).
		Run()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// selectPublishMode asks which trusted-publishing permissions to grant. At least
// one must be selected.
func selectPublishMode() (allowPublish, allowStage bool, err error) {
	for {
		var modes []string
		err = huh.NewMultiSelect[string]().
			Title("Publishing permissions").
			Description("Pick one or both (space to toggle, enter to confirm). At least one is required.").
			Options(
				huh.NewOption("Allow publish", "publish"),
				huh.NewOption("Allow staged publish", "stage"),
			).
			Value(&modes).
			Run()
		if err != nil {
			return false, false, err
		}
		for _, m := range modes {
			switch m {
			case "publish":
				allowPublish = true
			case "stage":
				allowStage = true
			}
		}
		if allowPublish || allowStage {
			return allowPublish, allowStage, nil
		}
		fmt.Fprintln(os.Stderr, "Select at least one permission.")
	}
}

// confirm asks a yes/no question, defaulting to no.
func confirm(prompt string) (bool, error) {
	var b bool
	err := huh.NewConfirm().Title(prompt).Value(&b).Run()
	return b, err
}
