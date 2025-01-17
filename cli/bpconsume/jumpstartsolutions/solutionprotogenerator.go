package jumpstartsolutions

import (
	"fmt"
	"strconv"
	"strings"

	gen_protos "github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpconsume/jumpstartsolutions/gen-protos"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpmetadata"
)

var RequiredRoles = []string{"roles/serviceusage.serviceUsageAdmin", "roles/iam.serviceAccountAdmin", "roles/resourcemanager.projectIamAdmin"}
var RequiredApis = []string{"config.googleapis.com"}
var DefaultInputs = []string{"project_id", "region", "labels"}

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
	addOutputs(solution, bpObj, bpDpObj)
	addDocumentationLink(solution, bpObj)
	addIsSingleton(solution, bpObj)
	addOrgPolicyChecks(solution, bpObj)
	addCloudProductIdentifiers(solution, bpObj)

	addIconUrl(solution)
	addDiagramUrl(solution)
	if err := validateSolutionProto(solution); err != nil {
		return nil, err
	}

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

	gitRepoSplit := strings.Split(solution.GitSource.Repo, "/")
	gitRepoName := gitRepoSplit[len(gitRepoSplit)-1]

	bpPathSplit := strings.Split(jssConsumptionFlags.bpPath, "/")
	var solutionModulePath string

	for i, dir := range bpPathSplit {
		if dir == gitRepoName {
			solutionModulePath = strings.Join(bpPathSplit[i+1:], "/")
			break
		}
	}

	solution.GitSource.Directory = solutionModulePath
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
	solution.CostEstimateUsd = float64(-1)
	if strings.HasPrefix(bpObj.Spec.Info.CostEstimate.Description, "cost of this solution is $") {
		solution.CostEstimateUsd, _ = strconv.ParseFloat(strings.TrimPrefix(bpObj.Spec.Info.CostEstimate.Description, "cost of this solution is $"), 64)
	}
}

func containsInList(element string, list []string) bool {
	for _, member := range list {
		if member == element {
			return true
		}
	}
	return false
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
	for _, role := range RequiredRoles {
		if !containsInList(role, solution.DeployData.Roles) {
			solution.DeployData.Roles = append(solution.DeployData.Roles, role)
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
	for _, api := range RequiredApis {
		if !containsInList(api, solution.DeployData.Apis) {
			solution.DeployData.Apis = append(solution.DeployData.Apis, api)
		}
	}
}

// addVariables adds terraform input variables to the solution object from
// the BlueprintMetadata object.
func addVariables(solution *gen_protos.Solution, bpObj, bpDpObj *bpmetadata.BlueprintMetadata) {
	if len(bpObj.Spec.Interfaces.Variables) == 0 {
		return
	}
	solution.DeployData.ConfigurationSections = []*gen_protos.ConfigurationSection{}
	for _, variable := range bpObj.Spec.Interfaces.Variables {
		if containsInList(variable.Name, DefaultInputs) || !variable.Required {
			//skipping the default and non-required inputs
			continue
		}
		bpVariable := bpDpObj.Spec.UI.Input.Variables[variable.Name]
		property := &gen_protos.ConfigurationProperty{
			Name:       variable.Name,
			IsRequired: variable.Required,
			IsHidden:   bpVariable.Invisible,
			Validation: bpVariable.RegExValidation,
		}
		switch variable.VarType {
		case "string":
			property.Type = gen_protos.ConfigurationProperty_STRING
			if variable.DefaultValue != nil {
				property.DefaultValue = fmt.Sprintf("%v", variable.DefaultValue)
			}
			property.Pattern = bpVariable.RegExValidation
			property.MaxLength = int32(bpVariable.Maximum)
			property.MinLength = int32(bpVariable.Minimum)

		case "bool":
			property.Type = gen_protos.ConfigurationProperty_BOOLEAN
			if variable.DefaultValue != nil {
				property.DefaultValue = fmt.Sprintf("%v", variable.DefaultValue)
			}
		case "list":
			property.Type = gen_protos.ConfigurationProperty_ARRAY
			property.MaxItems = int32(bpVariable.Maximum)
			property.MinItems = int32(bpVariable.Minimum)
		case "number":
			// Note: tf metadata uses "number" type for both "integer" and "number" type.
			// Hence, this might require manual update of textproto file.
			property.Type = gen_protos.ConfigurationProperty_NUMBER
			if variable.DefaultValue != nil {
				property.DefaultValue = fmt.Sprintf("%v", variable.DefaultValue)
			}
			property.Maximum = float32(bpVariable.Maximum)
			property.Minimum = float32(bpVariable.Minimum)
		}
		solution.DeployData.ConfigurationSections = append(solution.DeployData.ConfigurationSections, &gen_protos.ConfigurationSection{
			Properties: []*gen_protos.ConfigurationProperty{property},
		})
	}
}

// addOutputs adds terraform outputs to the solution object from the
// BlueprintMetadata object.
func addOutputs(solution *gen_protos.Solution, bpObj, bpDpObj *bpmetadata.BlueprintMetadata) {
	if len(bpObj.Spec.Interfaces.Outputs) == 0 {
		return
	}
	solution.DeployData.Links = []*gen_protos.DeploymentLink{}
	for _, link := range bpObj.Spec.Interfaces.Outputs {
		solutionOutput := bpDpObj.Spec.UI.Runtime.Outputs[link.Name]
		deploymentLink := &gen_protos.DeploymentLink{
			OutputName: link.Name,
		}
		if &solutionOutput != nil {
			deploymentLink.OpenInNewTab = solutionOutput.OpenInNewTab
			deploymentLink.ShowInNotification = solutionOutput.ShowInNotification
		}
		solution.DeployData.Links = append(solution.DeployData.Links, deploymentLink)
	}
}

// addDocumentationLink adds the URL of the solution's documentation page.
func addDocumentationLink(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	for _, documentation := range bpObj.Spec.Content.Documentation {
		if strings.ReplaceAll(strings.ToLower(documentation.Title), " ", "_") == "landing_page" {
			solution.DocumentationLink = documentation.URL
		} else if strings.ReplaceAll(strings.ToLower(documentation.Title), " ", "_") == "tutorial_walkthrough_id" {
			solution.NeosWalkthroughId = documentation.URL
		}
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

// addIsSingleton adds whether the solution is a singleton or not.
func addIsSingleton(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	solution.DeployData.IsSingleton = bpObj.Spec.Info.SingleDeployment
}

// addOrgPolicyChecks adds org policy checks to the solution object.
func addOrgPolicyChecks(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {
	solution.DeployData.OrgPolicyChecks = []*gen_protos.OrgPolicyCheck{}
	for _, orgPolicy := range bpObj.Spec.Info.OrgPolicyChecks {
		policy := &gen_protos.OrgPolicyCheck{
			Id:             orgPolicy.PolicyId,
			RequiredValues: orgPolicy.RequiredValues,
		}
		solution.DeployData.OrgPolicyChecks = append(solution.DeployData.OrgPolicyChecks, policy)
	}
}

// addCloudProductIdentifiers adds cloud product identifiers to the solution
// object.
func addCloudProductIdentifiers(solution *gen_protos.Solution, bpObj *bpmetadata.BlueprintMetadata) {

	solution.CloudProductIdentifiers = []*gen_protos.CloudProductIdentifier{}
	for _, cloudProduct := range bpObj.Spec.Info.CloudProducts {
		cpIdentifier := &gen_protos.CloudProductIdentifier{
			Label: cloudProduct.Label,
			ConsoleProductIdentifier: &gen_protos.ConsoleProductIdentifier{
				SectionId: cloudProduct.ProductId,
			},
		}
		if len(cloudProduct.PageURL) > 0 {
			if cloudProduct.IsExternal {
				cpIdentifier.ConsoleProductIdentifier.PageId = cloudProduct.PageURL
				cpIdentifier.ConsoleProductIdentifier.PageIdForPostDeploymentLink = cloudProduct.PageURL
			} else {
				cpIdentifier.ConsoleProductIdentifier.PageId = strings.ReplaceAll(cloudProduct.PageURL, "/", "_")
			}
		}
		solution.CloudProductIdentifiers = append(solution.CloudProductIdentifiers, cpIdentifier)
	}
	addLocationConfig(solution)
}
func addLocationConfig(solution *gen_protos.Solution) {
	for _, cloudProduct := range solution.CloudProductIdentifiers {
		switch cloudProduct.ConsoleProductIdentifier.SectionId {

		case "BIGQUERY_SECTION_transfers":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_BIGQUERY_DATA_TRANSFER)
			break
		case "CLOUD_BUILD_SECTION":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_CLOUD_BUILD)
			break
		case "CLOUD_DEPLOY_SECTION":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_CLOUD_DEPLOY)
			break
		case "FUNCTIONS_SECTION":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_CLOUD_FUNCTIONS_V2)
			break
		case "CACHE_SECTION":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_CLOUD_MEMORYSTORE)
			break
		case "SERVERLESS_SECTION":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_CLOUD_RUN)
			break
		case "COMPUTE_SECTION":
			solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_COMPUTE)
			break
		default:
			if strings.Contains(cloudProduct.ConsoleProductIdentifier.SectionId, "BIGQUERY") {
				solution.DeployData.LocationConfigs = append(solution.DeployData.LocationConfigs, gen_protos.DeployData_BIGQUERY_DATASET)
			}
			break
		}
	}

}
