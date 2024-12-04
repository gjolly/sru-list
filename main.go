package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
)

const AllTag = "ALL"

type Bug struct {
	URL string `yaml:"url"`
}

type SRU struct {
	Age             int    `yaml:"age"`
	Bugs            []*Bug `yaml:"bugs"`
	Package         string `yaml:"pkg"`
	ProposedVersion string `yaml:"proposed_version"`
	ReleaseVersion  string `yaml:"release_version"`
	UpdatesVersion  string `yaml:"updates_version"`
	Uploaders       string `yaml:"uploaders"`
	URL             string `yaml:"url"`
}

type Config struct {
	Packages       []string `yaml:"packages"`
	PackageRegexps []string `yaml:"package_regexps"`
}

func fetchSRUReport() (map[string][]*SRU, error) {
	client := new(http.Client)

	fileName := "sru_report.yaml"
	url := fmt.Sprintf("https://ubuntu-archive-team.ubuntu.com/%s", fileName)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read report: %w", err)
	}

	sruReports := make(map[string][]*SRU)
	err = yaml.Unmarshal(buffer.Bytes(), &sruReports)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML report: %w", err)
	}

	return sruReports, nil
}

func parseConfig(filename string) (map[string]struct{}, []*regexp.Regexp, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config: %w", err)
	}

	configContent := new(Config)
	err = yaml.Unmarshal(buffer.Bytes(), configContent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	packageMap := make(map[string]struct{})
	for _, pkg := range configContent.Packages {
		packageMap[pkg] = struct{}{}
	}

	packageRegexps := make([]*regexp.Regexp, len(configContent.PackageRegexps))
	for i, exp := range configContent.PackageRegexps {
		compliledExp, err := regexp.Compile(exp)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to compile regexp: %w", err)
		}
		packageRegexps[i] = compliledExp
	}

	return packageMap, packageRegexps, nil
}

func matchARegexp(pkg string, regexps []*regexp.Regexp) bool {
	for _, exp := range regexps {
		if exp.MatchString(pkg) {
			return true
		}
	}
	return false
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		slog.Error(fmt.Sprintf("Usage: %s CONFIG", flag.CommandLine.Name()))
		os.Exit(3)
	}

	packageList, pkgRegexps, err := parseConfig(flag.Arg(0))
	if err != nil {
		slog.Error("failed to parse config", "error", err)
		os.Exit(1)
	}

	report, err := fetchSRUReport()
	if err != nil {
		slog.Error("error fetching report", "error", err)
		os.Exit(2)
	}

	_, includeAll := packageList[AllTag]
	fmt.Println("package,release")
	for release, SRUs := range report {
		for _, sru := range SRUs {
			if _, ok := packageList[sru.Package]; includeAll || ok || matchARegexp(sru.Package, pkgRegexps) {
				fmt.Printf("%s,%s\n", sru.Package, release)
			}
		}
	}
}
