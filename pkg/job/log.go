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

package job

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// TailOptions configures log tailing behavior
type TailOptions struct {
	// Lines specifies the number of lines to tail (default: 5)
	Lines int

	// Filter is a regex pattern to filter log lines
	Filter string
}

// TailLog returns the last N lines from a job's log file
func (m *Manager) TailLog(jobID string, opts TailOptions) ([]string, error) {
	job, err := m.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	// Set default lines if not specified
	if opts.Lines <= 0 {
		opts.Lines = 5
	}

	// Read log file
	file, err := os.Open(job.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Apply filter if specified
	if opts.Filter != "" {
		lines, err = filterLines(lines, opts.Filter)
		if err != nil {
			return nil, fmt.Errorf("error applying filter: %w", err)
		}
	}

	// Return last N lines
	if len(lines) <= opts.Lines {
		return lines, nil
	}

	return lines[len(lines)-opts.Lines:], nil
}

// filterLines filters lines based on a regex pattern
func filterLines(lines []string, pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var filtered []string
	for _, line := range lines {
		if re.MatchString(line) {
			filtered = append(filtered, line)
		}
	}

	return filtered, nil
}

// ReadFullLog returns the entire log file content
func (m *Manager) ReadFullLog(jobID string) (string, error) {
	job, err := m.GetJob(jobID)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(job.LogPath)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %w", err)
	}

	return string(content), nil
}

// SearchLog searches for a pattern in the job's log file
func (m *Manager) SearchLog(jobID string, pattern string) ([]string, error) {
	job, err := m.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	// Read log file
	file, err := os.Open(job.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Search for matching lines
	var matches []string
	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, fmt.Sprintf("%d: %s", lineNum, line))
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return matches, nil
}

// GetLogStats returns statistics about a job's log file
func (m *Manager) GetLogStats(jobID string) (*LogStats, error) {
	job, err := m.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	// Get file info
	fileInfo, err := os.Stat(job.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	// Read and count lines
	file, err := os.Open(job.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var lineCount, wordCount, charCount int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		charCount += len(line)
		wordCount += len(strings.Fields(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return &LogStats{
		Path:      job.LogPath,
		Size:      fileInfo.Size(),
		Lines:     lineCount,
		Words:     wordCount,
		Characters: charCount,
		Modified:  fileInfo.ModTime(),
	}, nil
}

// LogStats contains statistics about a log file
type LogStats struct {
	Path       string
	Size       int64
	Lines      int
	Words      int
	Characters int
	Modified   time.Time
}
