package jumpstartsolutions

import (
	"fmt"
	"strings"

	gen_protos "github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpconsume/jumpstartsolutions/gen-protos"
)

func validateTextFields(textFields *JSSTextFields) error {
	// check null or empty
	var invalidFields []string
	if textFields == nil {
		return fmt.Errorf("unable to define text metadata for given solution")
	}
	if len(textFields.solutionName) == 0 {
		invalidFields = append(invalidFields, "solutionName")
	}
	if len(textFields.solutionDescription) == 0 {
		invalidFields = append(invalidFields, "solutionDescription")
	}
	if len(textFields.solutionTitle) == 0 {
		invalidFields = append(invalidFields, "solutionTitle")
	}
	if len(textFields.solutionSummary) == 0 {
		invalidFields = append(invalidFields, "solutionSummary")
	}
	if len(textFields.solutionDiagramSteps) == 0 {
		invalidFields = append(invalidFields, "solutionDiagramSteps")
	}
	if len(invalidFields) > 0 {
		return fmt.Errorf("These fields are missing or empty : " + strings.Join(invalidFields[:], ","))
	}
	return nil
}

func validateSolutionProto(solution *gen_protos.Solution) error {
	var invalidFields []string
	if solution == nil {
		return fmt.Errorf("unable to define service metadata for given solution")
	}
	if solution.GitSource == nil || len(solution.GitSource.Repo) == 0 {
		invalidFields = append(invalidFields, "Repository source")
	}
	if solution.CloudProductIdentifiers == nil || len(solution.CloudProductIdentifiers) == 0 {
		invalidFields = append(invalidFields, "CloudProductIdentifier")
	}
	if solution.DeploymentEstimate == nil {
		invalidFields = append(invalidFields, "DeploymentEstimate")
	}
	if solution.DeployData == nil {
		invalidFields = append(invalidFields, "Inputs, Outputs, Roles, APIs")
	} else {
		if len(solution.DeployData.Roles) == 0 {
			invalidFields = append(invalidFields, "Roles")
		}
		if len(solution.DeployData.Apis) == 0 {
			invalidFields = append(invalidFields, "APIs")
		}
		if len(solution.DeployData.Links) == 0 {
			invalidFields = append(invalidFields, "Outputs")
		}
		if len(solution.DeployData.ConfigurationSections) == 0 {
			invalidFields = append(invalidFields, "Inputs")
		}
	}
	if len(invalidFields) > 0 {
		return fmt.Errorf("These fields are missing or empty : " + strings.Join(invalidFields[:], ","))
	}
	return nil
}
