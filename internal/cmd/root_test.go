/*
MCP Shell Server for serving shell AI models
Copyright (C) 2025 Rafael

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cmd

import (
	"testing"
)

func TestSetVersion(t *testing.T) {
	testVersion := "1.0.0-test"
	SetVersion(testVersion)

	if version != testVersion {
		t.Errorf("Expected version to be %s, got %s", testVersion, version)
	}

	if rootCmd.Version != testVersion {
		t.Errorf("Expected rootCmd.Version to be %s, got %s", testVersion, rootCmd.Version)
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute() doesn't panic
	// In a real application, you might want to test specific command behavior
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
	}()

	// Note: This doesn't actually execute the command as it would interfere with testing
	// In practice, you'd mock the command execution or test specific functions
}
