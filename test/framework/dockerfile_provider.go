// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	quarkusJVMApplicationBinarySuffix    = "-runner.jar"
	quarkusNativeApplicationBinarySuffix = "-runner"
	springBootApplicationBinarySuffix    = ".jar"
)

// KogitoApplicationDockerfileProvider is the API to provide Dockerfile content for image creation based on built project content
type KogitoApplicationDockerfileProvider interface {
	// GetDockerfileContent returns Dockerfile content for image creation
	GetDockerfileContent() (string, error)
}

type kogitoApplicationDockerfileProviderStruct struct {
	projectLocation         string
	imageName               string
	applicationBinarySuffix string
	libFolderNeeded         bool
}

var quarkusNonNativeKogitoApplicationDockerfileProvider = kogitoApplicationDockerfileProviderStruct{
	imageName:               infrastructure.KogitoRuntimeJVM,
	applicationBinarySuffix: quarkusJVMApplicationBinarySuffix,
	libFolderNeeded:         true,
}

var quarkusNativeKogitoApplicationDockerfileProvider = kogitoApplicationDockerfileProviderStruct{
	imageName:               infrastructure.KogitoRuntimeNative,
	applicationBinarySuffix: quarkusNativeApplicationBinarySuffix,
	libFolderNeeded:         false,
}

var springbootKogitoApplicationDockerfileProvider = kogitoApplicationDockerfileProviderStruct{
	imageName:               infrastructure.KogitoRuntimeJVM,
	applicationBinarySuffix: springBootApplicationBinarySuffix,
	libFolderNeeded:         false,
}

// GetKogitoApplicationDockerfileProvider returns KogitoApplicationDockerfileProvider based on project location
func GetKogitoApplicationDockerfileProvider(projectLocation string) KogitoApplicationDockerfileProvider {
	targetDir := projectLocation + "/target"
	dockerfileProvider := &springbootKogitoApplicationDockerfileProvider

	if fileWithSuffixExists(targetDir, quarkusNativeKogitoApplicationDockerfileProvider.applicationBinarySuffix) {
		dockerfileProvider = &quarkusNativeKogitoApplicationDockerfileProvider
	} else if fileWithSuffixExists(targetDir, quarkusNonNativeKogitoApplicationDockerfileProvider.applicationBinarySuffix) {
		dockerfileProvider = &quarkusNonNativeKogitoApplicationDockerfileProvider
	}

	dockerfileProvider.projectLocation = projectLocation
	return dockerfileProvider
}

func (dockerfileProvider *kogitoApplicationDockerfileProviderStruct) GetDockerfileContent() (string, error) {
	// Declare base image to build from
	dockerfileContent := fmt.Sprintf("FROM %s\n", GetBuildImage(dockerfileProvider.imageName))

	// Copy application binary into $KOGITO_HOME/bin
	applicationName := filepath.Base(dockerfileProvider.projectLocation)
	dockerfileContent += fmt.Sprintf("COPY target/%s%s $KOGITO_HOME/bin\n", applicationName, dockerfileProvider.applicationBinarySuffix)

	// Copy lib folder
	if dockerfileProvider.libFolderNeeded {
		dockerfileContent += "COPY target/lib $KOGITO_HOME/bin/lib\n"
	}

	return dockerfileContent, nil
}

func fileWithSuffixExists(scannedDirectory, fileSuffix string) bool {
	files, err := ioutil.ReadDir(scannedDirectory)
	if err != nil {
		panic(fmt.Errorf("Error reading directory %s: %v", scannedDirectory, err))
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), fileSuffix) {
			return true
		}
	}
	return false
}
