package jumpstartsolutions

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/cli/bpmetadata"
)

const (
	soyDiagramDescriptionMsg = "  {msg desc=\"Step $COUNT of $SOLUTION_NAME diagram description\"}\n    $SOLUTION_DIAGRAM_DESCRIPTION\n  {/msg}\n"
	soyLineSeparator         = "  {\\n}\n"
)

type JSSTextFields struct {
	solutionName               string
	solutionId                 string
	solutionTitle              string
	solutionSummary            string
	solutionDescription        string
	solutionDiagramSteps       []string
	solutionDiagramDescription string
}

func generateSolutionId(solutionName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(solutionName), "-", "_"), " ", "_")
}

func generateReadableSolutionId(solutionId string) string {
	return strings.ReplaceAll(solutionId, "_", "-")
}
func createDiagramDescription(steps []string, solutionName string) string {
	var buffer bytes.Buffer
	for iteration, step := range steps {
		if iteration > 0 {
			buffer.WriteString(soyLineSeparator)
		}
		buffer.WriteString(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(soyDiagramDescriptionMsg, "$SOLUTION_DIAGRAM_DESCRIPTION", step), "$SOLUTION_NAME", solutionName), "$COUNT", strconv.Itoa(iteration+1)))
	}
	return buffer.String()
}

func generateSoy(bpObj *bpmetadata.BlueprintMetadata) error {
	textFields := &JSSTextFields{}
	textFields.solutionName = bpObj.Spec.Info.Title
	textFields.solutionId = generateSolutionId(textFields.solutionName)
	ingestionSolutionId := generateReadableSolutionId(textFields.solutionId)
	textFields.solutionTitle = bpObj.Spec.Info.Title
	textFields.solutionSummary = bpObj.Spec.Info.Description.Tagline
	textFields.solutionDescription = bpObj.Spec.Info.Description.Detailed

	if len(bpObj.Spec.Info.Description.Architecture) == 0 {
		textFields.solutionDiagramSteps = bpObj.Spec.Content.Architecture.Description
	} else {
		textFields.solutionDiagramSteps = bpObj.Spec.Info.Description.Architecture
	}
	solutionDiagramDescription := createDiagramDescription(textFields.solutionDiagramSteps, textFields.solutionName)

	if err := validateTextFields(textFields); err != nil {
		return err
	}
	replacer := strings.NewReplacer("$INGESTION_ID", ingestionSolutionId, "$SOLUTION_ID", textFields.solutionId, "$SOLUTION_NAME", textFields.solutionName, "$SOLUTION_TITLE", textFields.solutionTitle, "$SOLUTION_SUMMARY", textFields.solutionSummary, "$SOLUTION_DESCRIPTION", textFields.solutionDescription, "$DIAGRAM_DESCRIPTION", solutionDiagramDescription)

	input, err := ioutil.ReadFile("soy_template.soy")
	if err != nil {
		return err
	}
	output := replacer.Replace(string(input))
	fileName := textFields.solutionId + ".soy"
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err = os.WriteFile(path.Join(currentDir, fileName), []byte(output), 0644); err != nil {
		return err
	}
	return nil
}
