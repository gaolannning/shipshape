package cli

import (
	"flag"
	"testing"

	rpcpb "github.com/google/shipshape/shipshape/proto/shipshape_rpc_proto"
)

var (
	// There are two ways to specify test flags when using Bazel:
	// 1) In the BUILD file with an args stanza in the _test rule.
	// 2) On the command line using --test_arg (i.e. bazel test --test_arg=-shipshape_test_docker_tag=TAG ...).
	//
	// As of 9 Oct 2015, there are multiple Bazel targets that use --shipshape_test_docker_tag (:shipshape_test_prod
	// and :shipshape_test_local) but there are no targets that set local Kythe.
	dockerTag  = flag.String("shipshape_test_docker_tag", "", "the docker tag for the images to use for testing")
	localKythe = flag.Bool("shipshape_test_local_kythe", false, "if true, don't pull the Kythe docker image")
)

func countFailures(resp rpcpb.ShipshapeResponse) int {
	failures := 0
	for _, analyzeResp := range resp.AnalyzeResponse {
		failures += len(analyzeResp.Failure)
	}
	return failures
}

func countNotes(resp rpcpb.ShipshapeResponse) int {
	notes := 0
	for _, analyzeResp := range resp.AnalyzeResponse {
		notes += len(analyzeResp.Note)
	}
	return notes
}

func countCategoryNotes(resp rpcpb.ShipshapeResponse, category string) int {
	notes := 0
	for _, analyzeResp := range resp.AnalyzeResponse {
		for _, note := range analyzeResp.Note {
			if *note.Category == category {
				notes += 1
			}
		}
	}
	return notes
}

func TestExternalAnalyzers(t *testing.T) {
	// Replaces part of the e2e test
	// Create a fake maven project with android failures

	// Run CLI using a .shipshape file
}

func TestBuiltInAnalyzersPreBuild(t *testing.T) {
	options := Options{
		File:                "shipshape/cli/testdata/workspace1",
		ThirdPartyAnalyzers: []string{},
		Build:               "",
		TriggerCats:         []string{"PostMessage", "JSHint", "go vet", "PyLint"},
		Dind:                false,
		Event:               DefaultEvent,
		Repo:                DefaultRepo,
		StayUp:              true,
		Tag:                 *dockerTag,
		LocalKythe:          *localKythe,
	}
	var allResponses rpcpb.ShipshapeResponse
	options.HandleResponse = func(shipshapeResp *rpcpb.ShipshapeResponse, _ string) error {
		allResponses.AnalyzeResponse = append(allResponses.AnalyzeResponse, shipshapeResp.AnalyzeResponse...)
		return nil
	}
	returnedNotesCount, err := New(options).Run()
	if err != nil {
		t.Fatal(err)
	}
	testName := "TestBuiltInAnalyzerPreBuild"

	if got, want := countFailures(allResponses), 0; got != want {
		t.Errorf("%v: Wrong number of failures; got %v, want %v (proto data: %v)", testName, got, want, allResponses)
	}
	if countedNotes := countNotes(allResponses); returnedNotesCount != countedNotes {
		t.Errorf("%v: Inconsistent note count: returned %v, counted %v (proto data: %v", testName, returnedNotesCount, countedNotes, allResponses)
	}
	if got, want := returnedNotesCount, 39; got != want {
		t.Errorf("%v: Wrong number of notes; got %v, want %v (proto data: %v)", testName, got, want, allResponses)
	}
	if got, want := countCategoryNotes(allResponses, "PostMessage"), 2; got != want {
		t.Errorf("%v: Wrong number of PostMessage notes; got %v, want %v (proto data: %v)", testName, got, want, allResponses)
	}
	if got, want := countCategoryNotes(allResponses, "JSHint"), 3; got != want {
		t.Errorf("%v: Wrong number of JSHint notes; got %v, want %v (proto data: %v)", testName, got, want, allResponses)
	}
	if got, want := countCategoryNotes(allResponses, "go vet"), 1; got != want {
		t.Errorf("%v: Wrong number of go vet notes; got %v, want %v (proto data: %v)", testName, got, want, allResponses)
	}
	if got, want := countCategoryNotes(allResponses, "PyLint"), 33; got != want {
		t.Errorf("%v: Wrong number of PyLint notes; got %v, want %v (proto data: %v)", testName, got, want, allResponses)
	}
}

func TestBuiltInAnalyzersPostBuild(t *testing.T) {
	// Replaces part of the e2e test
	// Test with a kythe maven build
	// PostMessage and ErrorProne
}

func TestStreamsMode(t *testing.T) {
	// Test whether it works in streams mode
	// Before creating this, ensure that streams mode
	// is actually still something we need to support.
}

func TestChangingDirectories(t *testing.T) {
	// Run on parent directory
	options := Options{
		File:                "shipshape/cli/testdata/workspace2",
		ThirdPartyAnalyzers: []string{},
		Build:               "",
		TriggerCats:         []string{"PostMessage", "JSHint", "go vet", "PyLint"},
		Dind:                false,
		Event:               DefaultEvent,
		Repo:                DefaultRepo,
		StayUp:              true,
		Tag:                 *dockerTag,
		LocalKythe:          *localKythe,
	}
	var allResponsesParent rpcpb.ShipshapeResponse
	options.HandleResponse = func(shipshapeResp *rpcpb.ShipshapeResponse, _ string) error {
		allResponsesParent.AnalyzeResponse =
			append(allResponsesParent.AnalyzeResponse, shipshapeResp.AnalyzeResponse...)
		return nil
	}
	if _, err := New(options).Run(); err != nil {
		t.Fatal(err)
	}
	testName := "TestChangingDirectories - ParentDir"
	if got, want := countFailures(allResponsesParent), 0; got != want {
		t.Errorf("%v: Wrong number of failures; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent)
	}
	if got, want := countCategoryNotes(allResponsesParent, "JSHint"), 3; got != want {
		t.Errorf("%v: Wrong number of JSHint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent)
	}
	if got, want := countCategoryNotes(allResponsesParent, "go vet"), 1; got != want {
		t.Errorf("%v: Wrong number of go vet notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent)
	}
	if got, want := countCategoryNotes(allResponsesParent, "PyLint"), 22; got != want {
		t.Errorf("%v: Wrong number of PyLint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent)
	}

	// Moving down, running on child directory
	options = Options{
		File:                "shipshape/cli/testdata/workspace2/subworkspace1",
		ThirdPartyAnalyzers: []string{},
		Build:               "",
		TriggerCats:         []string{"PostMessage", "JSHint", "go vet", "PyLint"},
		Dind:                false,
		Event:               DefaultEvent,
		Repo:                DefaultRepo,
		StayUp:              true,
		Tag:                 *dockerTag,
		LocalKythe:          *localKythe,
	}
	var allResponsesChild rpcpb.ShipshapeResponse
	options.HandleResponse = func(shipshapeResp *rpcpb.ShipshapeResponse, _ string) error {
		allResponsesChild.AnalyzeResponse =
			append(allResponsesChild.AnalyzeResponse, shipshapeResp.AnalyzeResponse...)
		return nil
	}
	if _, err := New(options).Run(); err != nil {
		t.Fatal(err)
	}
	testName = "TestChangingDirectories - ChildDir"
	if got, want := countFailures(allResponsesChild), 0; got != want {
		t.Errorf("%v: Wrong number of failures; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesChild)
	}
	if got, want := countCategoryNotes(allResponsesChild, "JSHint"), 0; got != want {
		t.Errorf("%v: Wrong number of JSHint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesChild)
	}
	if got, want := countCategoryNotes(allResponsesChild, "go vet"), 1; got != want {
		t.Errorf("%v: Wrong number of go vet notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesChild)
	}
	if got, want := countCategoryNotes(allResponsesChild, "PyLint"), 0; got != want {
		t.Errorf("%v: Wrong number of PyLint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesChild)
	}

	// Moving sideways, running on the sibling child directory
	options = Options{
		File:                "shipshape/cli/testdata/workspace2/subworkspace2",
		ThirdPartyAnalyzers: []string{},
		Build:               "",
		TriggerCats:         []string{"PostMessage", "JSHint", "go vet", "PyLint"},
		Dind:                false,
		Event:               DefaultEvent,
		Repo:                DefaultRepo,
		StayUp:              true,
		Tag:                 *dockerTag,
		LocalKythe:          *localKythe,
	}
	var allResponsesSibling rpcpb.ShipshapeResponse
	options.HandleResponse = func(shipshapeResp *rpcpb.ShipshapeResponse, _ string) error {
		allResponsesSibling.AnalyzeResponse =
			append(allResponsesSibling.AnalyzeResponse, shipshapeResp.AnalyzeResponse...)
		return nil
	}
	if _, err := New(options).Run(); err != nil {
		t.Fatal(err)
	}
	testName = "TestChangingDirectories - SiblingDir"
	if got, want := countFailures(allResponsesSibling), 0; got != want {
		t.Errorf("%v: Wrong number of failures; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesSibling)
	}
	if got, want := countCategoryNotes(allResponsesSibling, "JSHint"), 0; got != want {
		t.Errorf("%v: Wrong number of JSHint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesSibling)
	}
	if got, want := countCategoryNotes(allResponsesSibling, "go vet"), 0; got != want {
		t.Errorf("%v: Wrong number of go vet notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesSibling)
	}
	if got, want := countCategoryNotes(allResponsesSibling, "PyLint"), 22; got != want {
		t.Errorf("%v: Wrong number of PyLint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesSibling)
	}

	// Moving up, run on the parent directory again
	options = Options{
		File:                "shipshape/cli/testdata/workspace2",
		ThirdPartyAnalyzers: []string{},
		Build:               "",
		TriggerCats:         []string{"PostMessage", "JSHint", "go vet", "PyLint"},
		Dind:                false,
		Event:               DefaultEvent,
		Repo:                DefaultRepo,
		StayUp:              true,
		Tag:                 *dockerTag,
		LocalKythe:          *localKythe,
	}
	var allResponsesParent2 rpcpb.ShipshapeResponse
	options.HandleResponse = func(shipshapeResp *rpcpb.ShipshapeResponse, _ string) error {
		allResponsesParent2.AnalyzeResponse =
			append(allResponsesParent2.AnalyzeResponse, shipshapeResp.AnalyzeResponse...)
		return nil
	}
	if _, err := New(options).Run(); err != nil {
		t.Fatal(err)
	}
	testName = "TestChangingDirectories - ParentDir again"
	if got, want := countFailures(allResponsesParent2), 0; got != want {
		t.Errorf("%v: Wrong number of failures; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent2)
	}
	if got, want := countCategoryNotes(allResponsesParent2, "JSHint"), 3; got != want {
		t.Errorf("%v: Wrong number of JSHint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent2)
	}
	if got, want := countCategoryNotes(allResponsesParent2, "go vet"), 1; got != want {
		t.Errorf("%v: Wrong number of go vet notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent2)
	}
	if got, want := countCategoryNotes(allResponsesParent2, "PyLint"), 22; got != want {
		t.Errorf("%v: Wrong number of PyLint notes; got %v, want %v (proto data: %v)",
			testName, got, want, allResponsesParent2)
	}

	// TODO(emso): Add a test running on a single file, use workspace2/test.js
}

func dumpLogs() {

}

func checkOutput(category string, numResults int) {

}
