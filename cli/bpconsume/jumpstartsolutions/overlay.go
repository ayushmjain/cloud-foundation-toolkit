package jumpstartsolutions

import (
	"os"

	gen_protos "github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpconsume/jumpstartsolutions/gen-protos"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/prototext"
)

// addOverlay reads the overlay.textproto file and overrides the given
// fields in the solution object.
func addOverlay(solution *gen_protos.Solution) error {
	b, err := os.ReadFile("overlay.textproto")

	// don't override the solution object if the overlay file doesn't exist
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	unmarshalOptions := prototext.UnmarshalOptions{}
	overlaySolution := &gen_protos.Solution{}

	err = unmarshalOptions.Unmarshal(b, overlaySolution)
	if err != nil {
		return err
	}

	overrideOutputs(solution, overlaySolution)
	overrideLocationConfigs(solution, overlaySolution)
	overrideCostEstimate(solution, overlaySolution)
	overrideGitSource(solution, overlaySolution)

	return nil
}

// overrideOutputs overrides the output fields of the solution object as per
// the overlaySolution object.
func overrideOutputs(solution, overlaySolution *gen_protos.Solution) {
	if overlaySolution.DeployData == nil {
		return
	}
	for _, overlayLink := range overlaySolution.DeployData.Links {
		if overlayLink.OutputName == "" {
			continue
		}
		for _, solutionLink := range solution.DeployData.Links {
			if overlayLink.OutputName == solutionLink.OutputName {
				solutionLink.OpenInNewTab = overlayLink.OpenInNewTab
				solutionLink.ShowInNotification = overlayLink.ShowInNotification
			}
		}
	}
}

// overrideLocationConfigs overrides the location configs fields of the
// solution object as per the overlaySolution object.
func overrideLocationConfigs(solution, overlaySolution *gen_protos.Solution) {
	if overlaySolution.DeployData == nil {
		return
	}
	if len(overlaySolution.DeployData.LocationConfigs) > 0 {
		solution.DeployData.LocationConfigs = append([]gen_protos.DeployData_DeployLocationConfig{}, overlaySolution.DeployData.LocationConfigs...)
	}

}

// overrideCostEstimate overrides the cost estimate USD field of the solution
// object as per the overlaySolution object.
func overrideCostEstimate(solution, overlaySolution *gen_protos.Solution) {
	if overlaySolution.CostEstimateUsd > 0 {
		solution.CostEstimateUsd = overlaySolution.CostEstimateUsd
	}
}

// overrideGitSource overrides the Git source field of the solution
// object as per the overlaySolution object.
func overrideGitSource(solution, overlaySolution *gen_protos.Solution) {
	if overlaySolution.GitSource == nil {
		return
	}
	if len(overlaySolution.GitSource.Ref) > 0 {
		solution.GitSource.Ref = overlaySolution.GitSource.Ref
	}
	if len(overlaySolution.GitSource.Directory) > 0 {
		solution.GitSource.Directory = overlaySolution.GitSource.Directory
	}
}
