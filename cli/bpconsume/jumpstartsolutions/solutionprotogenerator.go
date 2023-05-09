package jumpstartsolutions

import (
	"fmt"
	"strings"

	gen_protos "github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpconsume/jumpstartsolutions/gen-protos"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpmetadata"
)

// generateSolutionProto creates the Solution object from the BlueprintMetadata
// object.
func generateSolutionProto(bpObj, bpDpObj *bpmetadata.BlueprintMetadata) (*gen_protos.Solution, error) {
	solution := &gen_protos.Solution{}

	addGitSource(solution, bpObj)
	addDeploymentTimeEstimate(solution, bpObj)
	addCostEstimate(solution, bpObj)

	solution.DeployData = &gen_protos.DeployData{}
	err := addRoles(solution, bpObj)
	if err != nil {
		return nil, err
	}

	addApis(solution, bpObj)
	addVariables(solution, bpObj, bpDpObj)
	addOutputs(solution, bpObj)

	addIconUrl(solution)
	addDiagramUrl(solution)
	addDocumentationLink(solution)
	addIsSingleton(solution)
	addLocationConfigs(solution)
	addOrgPolicyChecks(solution)
	addCloudProductIdentifiers(solution)

	return solution, nil
}

// addGitSource adds the solution's git source to the solution object from the
// BlueprintMetadata object.
func addGitSource(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	if solution.GitSource == nil {
		solution.GitSource = &gen_protos.GitSource{}
	}
	if bpObj.Spec.Info.Source != nil {
		solution.GitSource.Repo = strings.TrimSuffix(bpObj.Spec.Info.Source.Repo, ".git")
	}

	// Placeholders for fields that aren't available in OSS metadata
	solution.GitSource.Ref = "<Git branch or tag or commit hash>"
	solution.GitSource.Directory = "<Subdirectory inside the repository>"
}

// addDeploymentTimeEstimate adds the deployment time for the solution to the
// solution object from the BlueprintMetadata object.
func addDeploymentTimeEstimate(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	if bpObj.Spec.Info.DeploymentDuration.ConfigurationSecs > 0 && bpObj.Spec.Info.DeploymentDuration.DeploymentSecs > 0 {
		solution.DeploymentEstimate = &gen_protos.DeploymentEstimate{
			// adding 59 (60 - 1) so that the result is ceiling after division.
			// Using fast ceiling of integer division method.
			ConfigurationMinutes: int32((bpObj.Spec.Info.DeploymentDuration.ConfigurationSecs + 59) / 60),
			DeploymentMinutes:    int32((bpObj.Spec.Info.DeploymentDuration.DeploymentSecs + 59) / 60),
		}
	}
}

// addCostEstimate adds the cost estimate for the solution to the solution
// object from the BlueprintMetadata object.
func addCostEstimate(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	if bpObj.Spec.Info.CostEstimate.URL != "" {
		solution.CostEstimateLink = bpObj.Spec.Info.CostEstimate.URL
	}
	solution.CostEstimateUsd = 1.0
}

// addRoles adds the roles required by the service account deploying the
// solution to the solution object from the BlueprintMetdata object.
func addRoles(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) error {
	if len(bpObj.Spec.Requirements.Roles) == 0 {
		return nil
	}
	projectRoleCount := 0
	for _, bpRoles := range bpObj.Spec.Requirements.Roles {
		if bpRoles.Level == "Project" {
			projectRoleCount += 1
		}
	}
	if projectRoleCount > 1 {
		return fmt.Errorf("more than one set of project level roles present in OSS solution metadata")
	}
	for _, bpRoles := range bpObj.Spec.Requirements.Roles {
		if bpRoles.Level == "Project" {
			solution.DeployData.Roles = make([]string, len(bpRoles.Roles))
			copy(solution.DeployData.Roles, bpRoles.Roles)
		}
	}
	return nil
}

// addApis adds the APIs required for deploying the solution to the solution
// object from the BlueprintMetadata object.
func addApis(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	if len(bpObj.Spec.Requirements.Services) == 0 {
		return
	}
	solution.DeployData.Apis = make([]string, len(bpObj.Spec.Requirements.Services))
	copy(solution.DeployData.Apis, bpObj.Spec.Requirements.Services)
}

// addVariables adds terraform input variables to the solution object from
// the BlueprintMetadata object.
func addVariables(solution *gen_protos.Solution, bpObj, bpDpObj *bpmetadata.BlueprintMetadata) {
	if len(bpObj.Spec.Interfaces.Variables) == 0 {
		return
	}
	solution.DeployData.InputSections = []*gen_protos.Section{}
	for _, variable := range bpObj.Spec.Interfaces.Variables {
		bpVariable := bpDpObj.Spec.UI.Input.Variables[variable.Name]
		property := &gen_protos.Property{
			Name:       variable.Name,
			IsRequired: variable.Required,
			IsHidden:   bpVariable.Invisible,
			Validation: bpVariable.RegExValidation,
		}
		switch variable.VarType {
		case "string":
			property.Type = gen_protos.Property_STRING
			if variable.DefaultValue != nil {
				property.DefaultValue = fmt.Sprintf("%v", variable.DefaultValue)
			}
			property.Pattern = bpVariable.RegExValidation
			property.MaxLength = int32(bpVariable.Maximum)
			property.MinLength = int32(bpVariable.Minimum)

		case "bool":
			property.Type = gen_protos.Property_BOOLEAN
			if variable.DefaultValue != nil {
				property.DefaultValue = fmt.Sprintf("%v", variable.DefaultValue)
			}
		case "list":
			property.Type = gen_protos.Property_ARRAY
			property.MaxItems = int32(bpVariable.Maximum)
			property.MinItems = int32(bpVariable.Minimum)
		case "number":
			// Note: tf metadata uses "number" type for both "integer" and "number" type.
			// Hence, this might require manual update of textproto file.
			property.Type = gen_protos.Property_NUMBER
			if variable.DefaultValue != nil {
				property.DefaultValue = fmt.Sprintf("%v", variable.DefaultValue)
			}
			property.Maximum = float32(bpVariable.Maximum)
			property.Minimum = float32(bpVariable.Minimum)
		}
		solution.DeployData.InputSections = append(solution.DeployData.InputSections, &gen_protos.Section{
			Properties: []*gen_protos.Property{property},
		})
	}
}

// addOutputs adds terraform outputs to the solution object from the
// BlueprintMetadata object.
func addOutputs(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	if len(bpObj.Spec.Interfaces.Outputs) == 0 {
		return
	}
	solution.DeployData.Links = []*gen_protos.DeploymentLink{}
	for _, link := range bpObj.Spec.Interfaces.Outputs {
		solution.DeployData.Links = append(solution.DeployData.Links, &gen_protos.DeploymentLink{
			OutputName: link.Name,
		})
	}
}

// addIconUrl adds the URL of the solution's icon image.
func addIconUrl(solution *gen_protos.Solution) {
	solution.IconUrl = "solution_icon.png"
}

// addDiagramUrl adds the URL of the solution's architecture diagram image.
func addDiagramUrl(solution *gen_protos.Solution) {
	solution.DiagramUrl = "solution_diagram.png"
}

// addDocumentationLink adds the URL of the solution's documentation page.
func addDocumentationLink(solution *gen_protos.Solution) {
	solution.DocumentationLink = "<cloud documentation link for the solution e.g. https://cloud.google.com/architecture/big-data-analytics/data-warehouse>"
}

// addIsSingleton adds whether the solution is a singleton or not.
func addIsSingleton(solution *gen_protos.Solution) {
	solution.DeployData.IsSingleton = true
}

// addLocationConfigs adds location configs to the solution object.
func addLocationConfigs(solution *gen_protos.Solution) {
	solution.DeployData.LocationConfigs = []gen_protos.DeployData_DeployLocationConfig{gen_protos.DeployData_UNSPECIFIED}
}

// addOrgPolicyChecks adds org policy checks to the solution object.
func addOrgPolicyChecks(solution *gen_protos.Solution) {
	solution.DeployData.OrgPolicyChecks = []*gen_protos.OrgPolicyCheck{{
		Id:             "<Org policy constraint e.g. constraints/gcp.resourceLocations>",
		RequiredValues: []string{"<required value 1>", "<required value 2>"},
	}}
}

// addCloudProductIdentifiers adds cloud product identifiers to the solution
// object.
func addCloudProductIdentifiers(solution *gen_protos.Solution) {
	solution.CloudProductIdentifiers = []*gen_protos.CloudProductIdentifier{{
		Label: "<product label>",
		ConsoleProductIdentifier: &gen_protos.ConsoleProductIdentifier{
			SectionId:                   "<product section ID>",
			PageId:                      "<product page ID>",
			PageIdForPostDeploymentLink: "<product page ID for post deployment link>",
		},
	}}
}
