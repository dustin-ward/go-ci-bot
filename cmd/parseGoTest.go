package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type JsonData struct {
	Time    string
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

func (jd *JsonData) UnmarshalJSON(b []byte) error {
	type Alias JsonData
	type Aux struct {
		Test    *string  `json:"Test"`
		Output  *string  `json:"Output"`
		Elapsed *float64 `json:"Elapsed"`
		*Alias
	}
	aux := &Aux{Alias: (*Alias)(jd)}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	if aux.Test == nil {
		jd.Test = ""
	} else {
		jd.Test = *aux.Test
	}

	if aux.Output == nil {
		jd.Output = ""
	} else {
		jd.Output = *aux.Output
	}

	if aux.Elapsed == nil {
		jd.Elapsed = -1
	} else {
		jd.Elapsed = *aux.Elapsed
	}
	return nil
}

type TestSummary struct {
	Pass  int
	Fail  int
	Skip  int
	Total int

	// each package result: package name -> the result of the package
	PackageResults map[string]*PackageResult
}

func (ts *TestSummary) IsAllPass() bool {
	return ts.Pass == ts.Total
}

func (ts *TestSummary) HasFail() bool {
	return ts.Fail > 0
}

// return package names that the test runned
func (ts *TestSummary) TestPackageList() []string {
	result := make([]string, 0, 3)
	for pkgName, _ := range ts.PackageResults {
		result = append(result, pkgName)
	}
	return result
}

func (ts *TestSummary) String() string {
	return fmt.Sprintf("total: %d, pass: %d, fail: %d, skip: %d", ts.Total, ts.Pass, ts.Fail, ts.Skip)
}

type PackageResult struct {
	Package    string
	RunTests   []string          // names of the tests that are runned. This will be to total tests that are runned
	PassTests  []string          // names of the tests that are passed
	FailTests  []string          // names of the tests that are failed
	SkipTests  []string          // names of the tests that are skiped
	TestOutput map[string]string // output of each test results: test name -> testcase output
}

func NewPackageResult(pkgName string) *PackageResult {
	pr := new(PackageResult)
	pr.Package = pkgName
	pr.RunTests = make([]string, 0, 10)
	pr.PassTests = make([]string, 0, 10)
	pr.FailTests = make([]string, 0, 10)
	pr.SkipTests = make([]string, 0, 10)
	pr.TestOutput = make(map[string]string)
	return pr
}

func (pr *PackageResult) IsAllPass() bool {
	return len(pr.RunTests) == len(pr.PassTests)
}

func (pr *PackageResult) HasFail() bool {
	return len(pr.FailTests) > 0
}

func (pr *PackageResult) Summary() (output string) {
	run := len(pr.RunTests)
	pass := len(pr.PassTests)
	fail := len(pr.FailTests)
	skip := len(pr.SkipTests)
	output = fmt.Sprintf("package: %s, run: %d, pass: %d, fail: %d, skip: %d",
		pr.Package,
		run,
		pass,
		fail,
		skip)
	return
}

func (pr *PackageResult) AllTestStatus() map[string]string {
	result := make(map[string]string)
	// init all test to unknow status
	// e.g: setting "testA": "?"
	for _, runTest := range pr.RunTests {
		result[runTest] = "?"
	}

	// e.g: setting "testA": "pass | fail | skip"
	setTestStatus(&result, &pr.PassTests, "pass")
	setTestStatus(&result, &pr.FailTests, "fail")
	setTestStatus(&result, &pr.SkipTests, "skip")
	return result
}

func setTestStatus(testSet *map[string]string, statusTestSet *[]string, status string) {
	// testSet must be init first
	for _, test := range *statusTestSet {
		if _, ok := (*testSet)[test]; ok {
			(*testSet)[test] = status
		} else {
			(*testSet)[test] = "? (test is not in runned test set)"
		}
	}
}

func (pr *PackageResult) String() string {
	summaryStr := pr.Summary()
	var sb strings.Builder
	sb.WriteString(summaryStr)

	allTestStatus := pr.AllTestStatus()
	for testName, testStatus := range allTestStatus {
		// output:
		// \t testName: ok | fail | skip | ?
		// "?" for test in unknow result
		sb.WriteString("\t")
		sb.WriteString(testName)
		sb.WriteString(": ")
		sb.WriteString(testStatus)
		sb.WriteString("\n")
	}
	return sb.String()

}

func parseGoTestJson(text string) (*TestSummary, error) {
	summary := new(TestSummary)
	summary.PackageResults = make(map[string]*PackageResult)
	scanner := bufio.NewScanner(strings.NewReader(text))
	var errs []error

	for scanner.Scan() {
		var data JsonData = JsonData{}
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			parseErr := fmt.Errorf("parsing error: %s in line: %s\n", err.Error(), line)
			errs = append(errs, parseErr)
			continue
		}

		pkgName := data.Package
		if pkgName == "" {
			println("package name is not string")
			continue
		}

		if _, ok := summary.PackageResults[pkgName]; !ok {
			summary.PackageResults[pkgName] = NewPackageResult(pkgName)
		}

		if test := data.Test; test != "" {
			// collect the test output data
			if output := data.Output; output != "" {
				summary.PackageResults[pkgName].TestOutput[test] += output
			} else if elapsed := data.Elapsed; elapsed >= 0 {
				summary.PackageResults[pkgName].TestOutput[test] += "Elapsed: " + strconv.FormatFloat(elapsed, 'g', -1, 64) + "\n"
			} else {
				if _, ok := summary.PackageResults[pkgName].TestOutput[test]; ok {
					summary.PackageResults[pkgName].TestOutput[test] += test + "\n"
				} else {
					summary.PackageResults[pkgName].TestOutput[test] = test + "\n"
				}
			}

			// depend on the action record the test
			if action := data.Action; action != "" {
				switch action {
				case "output", "pause", "cont":
					// do nothing. output is handle by above
				case "run":
					summary.Total++
					summary.PackageResults[pkgName].RunTests = append(summary.PackageResults[pkgName].RunTests, test)
				case "pass":
					summary.Pass++
					summary.PackageResults[pkgName].PassTests = append(summary.PackageResults[pkgName].PassTests, test)
				case "fail":
					summary.Fail++
					summary.PackageResults[pkgName].FailTests = append(summary.PackageResults[pkgName].FailTests, test)
				case "skip":
					summary.Skip++
					summary.PackageResults[pkgName].SkipTests = append(summary.PackageResults[pkgName].SkipTests, test)
				default:
					println("unknow action: ", action)
				}
			}

		}
	}

	if len(errs) > 0 {
		var str string
		for _, err := range errs {
			str += err.Error() + "\n"
		}
		return summary, fmt.Errorf("%s", str)
	}

	return summary, nil
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	return fileInfo.IsDir()
}

func readFile(path string) string {
	text, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(text)
}

func readInput(args []string) string {
	currentPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var sb strings.Builder
	var targetPath string
	for _, arg := range args {
		if filepath.IsAbs(arg) {
			targetPath = arg 
		} else {
			targetPath = filepath.Join(currentPath, arg)
		}

		if isDir(targetPath) {
			files, err := os.ReadDir(targetPath)
			if err != nil {
				panic(err)
			}

			for _, fileEntry := range files {
				filePath := filepath.Join(targetPath, fileEntry.Name())
				if fileEntry.IsDir() {
					continue
				}
				if !(strings.HasSuffix(filePath, ".testout") || strings.HasSuffix(filePath, ".tmp")) {
					continue
				}

				sb.WriteString(readFile(filePath))
			}
		} else { // file is not a directory
			sb.WriteString(readFile(targetPath))
		}
	}
	return sb.String()
}

func main() {

	flag.Usage = func() {
		helpText := `the program parse the output of go test -json.
		parse-go-test [files] [directories]`

		fmt.Println(helpText)
	}

	flag.Parse()

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("please pass a file or directory as arguments")
		return
	}

	testoutput := readInput(args)

	goTestSummary, parseError := parseGoTestJson(testoutput)
	if parseError != nil {
		println("Cannot parse json data: \n", parseError.Error())
	}

	fmt.Println("Test Summary")
	fmt.Println(goTestSummary)
	fmt.Println("\n")

	packageList := goTestSummary.TestPackageList()

	fmt.Println("Fail Test:")
	for _, pkgName := range packageList {
		if goTestSummary.PackageResults[pkgName].HasFail() {
			fmt.Println("Fail on: ", goTestSummary.PackageResults[pkgName].Summary())
			fmt.Println("-----------------------------------------------------------")
			for _, failTestName := range goTestSummary.PackageResults[pkgName].FailTests {
				fmt.Println()
				fmt.Println(goTestSummary.PackageResults[pkgName].TestOutput[failTestName])
			}
			fmt.Println("-----------------------------------------------------------")
		}
	}

}

func exampleUsage() {

	testoutput := `{"Time":"2024-05-29T10:50:26.103587893-04:00","Action":"start","Package":"syscall"}
{"Time":"2024-05-29T10:50:26.231295252-04:00","Action":"run","Package":"syscall","Test":"TestUnixCredentials"}
{"Time":"2024-05-29T10:50:26.231321264-04:00","Action":"output","Package":"syscall","Test":"TestUnixCredentials","Output":"=== RUN   TestUnixCredentials\n"}
{"Time":"2024-05-29T10:50:26.231336089-04:00","Action":"output","Package":"syscall","Test":"TestUnixCredentials","Output":"--- PASS: TestUnixCredentials (0.00s)\n"}
{"Time":"2024-05-29T10:50:26.231341102-04:00","Action":"pass","Package":"syscall","Test":"TestUnixCredentials","Elapsed":0}
`
	goTestSummary, err := parseGoTestJson(testoutput)
	if err != nil {
		println("Cannot parse json data: ", err.Error())
		return
	}

	fmt.Println("Test Summary")
	fmt.Println(goTestSummary)
	fmt.Println("\n")

	fmt.Println("\n========================\n")
	packageList := goTestSummary.TestPackageList()
	fmt.Println("package test: ", packageList)

	for _, pkgName := range packageList {
		fmt.Println("\n----------------------------------\n")
		fmt.Println(goTestSummary.PackageResults[pkgName])
	}

	fmt.Println("\n==================================\n")

	fmt.Println("Fail test:")
	for _, pkgName := range packageList {
		if goTestSummary.PackageResults[pkgName].HasFail() {
			fmt.Println(goTestSummary.PackageResults[pkgName].Summary())
			for _, failTestName := range goTestSummary.PackageResults[pkgName].FailTests {
				fmt.Println(failTestName)
				fmt.Println(goTestSummary.PackageResults[pkgName].TestOutput[failTestName])
			}
		}
	}
}
