/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusterconfig

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

func TestNetworkConfig(t *testing.T) {
	testcases := []struct {
		name                 string
		filesToCreate        []string
		ipAddressFileContent string
		expectedErr          bool
		validateFunc         func(t *testing.T, tmpDir string, err error, files []string, unc UpgradeNetworkConfigGather)
	}{
		{
			name: "Validate success flow",
			filesToCreate: []string{"/etc/hostname", "/etc/NetworkManager/system-connections/test1.txt",
				"/etc/NetworkManager/system-connections/scripts/test1.txt", ipAddressFile},
			expectedErr:          false,
			ipAddressFileContent: "192.168.127.10",
			validateFunc: func(t *testing.T, tmpDir string, err error, files []string, unc UpgradeNetworkConfigGather) {
				dir, err := unc.configDir(tmpDir)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				counter := 0
				for _, file := range files {
					err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
						if err == nil && info.Name() == filepath.Base(file) {
							counter++
						}
						return nil
					})
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				}
				assert.Equal(t, len(files), counter)
			},
		},
		{
			name: "No ip address file should fail",
			filesToCreate: []string{"/etc/hostname", "/etc/NetworkManager/system-connections/test1.txt",
				"/etc/NetworkManager/system-connections/scripts/test1.txt"},
			expectedErr:          true,
			ipAddressFileContent: "",
			validateFunc: func(t *testing.T, tmpDir string, err error, files []string, unc UpgradeNetworkConfigGather) {
				assert.Equal(t, err != nil, true)
			},
		},
		{
			name: "Bad ip address found, should fail",
			filesToCreate: []string{"/etc/hostname", "/etc/NetworkManager/system-connections/test1.txt",
				"/etc/NetworkManager/system-connections/scripts/test1.txt", ipAddressFile},
			expectedErr:          true,
			ipAddressFileContent: "bad ip",
			validateFunc: func(t *testing.T, tmpDir string, err error, files []string, unc UpgradeNetworkConfigGather) {
				assert.Equal(t, err != nil, true)
			},
		},
	}

	for _, tc := range testcases {
		tmpDir := t.TempDir()
		t.Run(tc.name, func(t *testing.T) {
			listOfPaths = []string{}
			hostPath = tmpDir
			// create list of files to copy
			for _, path := range tc.filesToCreate {
				dir := filepath.Join(tmpDir, filepath.Dir(path))
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				newPath := filepath.Join(dir, filepath.Base(path))
				f, err := os.Create(newPath)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if path == ipAddressFile {
					_, err = f.WriteString(tc.ipAddressFileContent)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				}

				_ = f.Close()
				listOfPaths = append(listOfPaths, path)
			}

			unc := UpgradeNetworkConfigGather{
				Log: logr.Discard(),
			}
			err := unc.FetchNetworkConfig(context.TODO(), tmpDir)
			if !tc.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErr && err == nil {
				t.Errorf("expected error but it didn't happened")
			}
			tc.validateFunc(t, tmpDir, err, tc.filesToCreate, unc)
		})
	}
}
