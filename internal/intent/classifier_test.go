package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPromptStrategies(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "prompt-strategies-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directories to test recursion
	nestedDir := filepath.Join(tmpDir, "nested")
	err = os.Mkdir(nestedDir, 0755)
	require.NoError(t, err)

	// Create test files with different extensions
	testFiles := map[string]string{
		filepath.Join(tmpDir, "strategy_one.md"):       "This is the first strategy.",
		filepath.Join(tmpDir, "strategy_two.md"):       "This is the second strategy.",
		filepath.Join(nestedDir, "nested_strategy.md"): "This is a nested strategy.",
		filepath.Join(tmpDir, "custom_strategy.txt"):   "This is a custom extension strategy.",
		filepath.Join(tmpDir, "not_included.json"):     "This should be ignored by default.",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test 1: Default extension (.md)
	result := LoadPromptStrategies(tmpDir, "")

	// Verify results for default extension
	assert.Contains(t, result, "# Strategy One")
	assert.Contains(t, result, "This is the first strategy.")
	assert.Contains(t, result, "# Strategy Two")
	assert.Contains(t, result, "This is the second strategy.")
	assert.Contains(t, result, "# Nested Strategy")
	assert.Contains(t, result, "This is a nested strategy.")
	assert.NotContains(t, result, "# Custom Strategy")
	assert.NotContains(t, result, "This is a custom extension strategy.")
	assert.NotContains(t, result, "# Not Included")
	assert.NotContains(t, result, "This should be ignored by default.")

	// Verify the separator is correctly added
	assert.Contains(t, result, "---")

	// Test 2: Custom extension (.txt)
	txtResult := LoadPromptStrategies(tmpDir, ".txt")
	assert.Contains(t, txtResult, "# Custom Strategy")
	assert.Contains(t, txtResult, "This is a custom extension strategy.")
	assert.NotContains(t, txtResult, "# Strategy One")
	assert.NotContains(t, txtResult, "This is the first strategy.")

	// Test 3: Custom extension without dot prefix
	txtResult2 := LoadPromptStrategies(tmpDir, "txt")
	assert.Contains(t, txtResult2, "# Custom Strategy")
	assert.Contains(t, txtResult2, "This is a custom extension strategy.")
	assert.NotContains(t, txtResult2, "# Strategy One")

	// Test 4: Non-existent directory should return default strategy
	nonExistentResult := LoadPromptStrategies("/this/directory/does/not/exist", "")
	assert.Equal(t, DefaultStrategy, nonExistentResult)

	// Test 5: Directory with no matching files should return default strategy
	emptyDir, err := os.MkdirTemp("", "empty-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(emptyDir)

	emptyResult := LoadPromptStrategies(emptyDir, "")
	assert.Equal(t, DefaultStrategy, emptyResult)

	// Test 6: Directory with files but none matching the extension
	noMatchResult := LoadPromptStrategies(tmpDir, ".xyz")
	assert.Equal(t, DefaultStrategy, noMatchResult)
}

func TestStrategyFormatting(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "format-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a file with a complex name to test formatting
	fileName := "complex_file_name_with_underscores.md"
	fileContent := "This content should be preceded by a properly formatted heading."
	filePath := filepath.Join(tmpDir, fileName)

	err = os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err)

	// Load the strategy
	result := LoadPromptStrategies(tmpDir, "")

	// Check that underscores were replaced with spaces and title case was applied
	assert.Contains(t, result, "# Complex File Name With Underscores")
	assert.True(t, strings.Contains(result, "This content should be preceded"))
}

func TestEmptyOrErrorFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create an empty file
	emptyPath := filepath.Join(tmpDir, "empty.md")
	err = os.WriteFile(emptyPath, []byte(""), 0644)
	require.NoError(t, err)

	// Create a directory with the .md extension (shouldn't be processed as a file)
	dirPath := filepath.Join(tmpDir, "directory.md")
	err = os.Mkdir(dirPath, 0755)
	require.NoError(t, err)

	// Load the strategies
	result := LoadPromptStrategies(tmpDir, "")

	// Should contain the empty file heading but with no content
	assert.Contains(t, result, "# Empty")

	// The directory.md should not be processed
	// We can verify this by checking if there's exactly one heading in the result
	headingCount := strings.Count(result, "# ")
	assert.Equal(t, 1, headingCount, "Should only contain one heading")
}

func TestLoadStrategyFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "load-strategies-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested directory
	nestedDir := filepath.Join(tmpDir, "nested")
	err = os.Mkdir(nestedDir, 0755)
	require.NoError(t, err)

	// Create test files
	testFiles := map[string]string{
		filepath.Join(tmpDir, "strategy_one.md"):       "Content one",
		filepath.Join(tmpDir, "strategy_two.md"):       "Content two",
		filepath.Join(nestedDir, "nested_strategy.md"): "Nested content",
		filepath.Join(tmpDir, "different.txt"):         "Different extension",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test with default extension
	contents, paths, err := LoadStrategyFiles(tmpDir, "")
	require.NoError(t, err)

	// Verify correct loading
	assert.Equal(t, 3, len(contents), "Should load 3 md files")
	assert.Equal(t, 3, len(paths), "Should have 3 paths")

	// Check contents and names
	assert.Equal(t, "Content one", contents["Strategy One"])
	assert.Equal(t, "Content two", contents["Strategy Two"])
	assert.Equal(t, "Nested content", contents["Nested Strategy"])

	// Check paths
	assert.Contains(t, paths["Strategy One"], "strategy_one.md")
	assert.Contains(t, paths["Strategy Two"], "strategy_two.md")
	assert.Contains(t, paths["Nested Strategy"], "nested_strategy.md")

	// Test with custom extension
	txtContents, txtPaths, err := LoadStrategyFiles(tmpDir, ".txt")
	require.NoError(t, err)

	assert.Equal(t, 1, len(txtContents), "Should load 1 txt file")
	assert.Equal(t, "Different extension", txtContents["Different"])
	assert.Contains(t, txtPaths["Different"], "different.txt")

	// Test with non-existent directory
	defaultContents, defaultPaths, err := LoadStrategyFiles("/non/existent/dir", "")
	require.NoError(t, err)

	assert.Equal(t, 1, len(defaultContents), "Should contain default strategy")
	assert.Equal(t, "built-in", defaultPaths["Default Strategy"])
	assert.Equal(t, DefaultStrategy, defaultContents["Default Strategy"])
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		vec1     map[string]float64
		vec2     map[string]float64
		expected float64
	}{
		{
			name:     "Identical vectors",
			vec1:     map[string]float64{"a": 1, "b": 2, "c": 3},
			vec2:     map[string]float64{"a": 1, "b": 2, "c": 3},
			expected: 1.0,
		},
		{
			name:     "Orthogonal vectors",
			vec1:     map[string]float64{"a": 1, "b": 0, "c": 0},
			vec2:     map[string]float64{"a": 0, "b": 1, "c": 0},
			expected: 0.0,
		},
		{
			name:     "Similar vectors",
			vec1:     map[string]float64{"a": 1, "b": 2, "c": 3},
			vec2:     map[string]float64{"a": 2, "b": 1, "c": 3},
			expected: 0.9286, // Rounded to 4 decimal places (actual: 0.9285714285714286)
		},
		{
			name:     "Empty vector 1",
			vec1:     map[string]float64{},
			vec2:     map[string]float64{"a": 1, "b": 2, "c": 3},
			expected: 0.0,
		},
		{
			name:     "Empty vector 2",
			vec1:     map[string]float64{"a": 1, "b": 2, "c": 3},
			vec2:     map[string]float64{},
			expected: 0.0,
		},
		{
			name:     "Disjoint vectors",
			vec1:     map[string]float64{"a": 1, "b": 2, "c": 3},
			vec2:     map[string]float64{"d": 1, "e": 2, "f": 3},
			expected: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CosineSimilarity(tc.vec1, tc.vec2)
			if tc.expected == 0 {
				assert.Equal(t, tc.expected, result)
			} else {
				assert.InDelta(t, tc.expected, result, 0.0001)
			}
		})
	}
}

func TestClassifyIntent(t *testing.T) {
	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "classify-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test strategy files
	testFiles := map[string]string{
		filepath.Join(tmpDir, "golang.md"):     "Go is a statically typed compiled language. It has garbage collection, limited structural typing, memory safety and goroutines for concurrency.",
		filepath.Join(tmpDir, "python.md"):     "Python is a high-level, interpreted, general-purpose programming language. Its design philosophy emphasizes code readability with the use of significant indentation.",
		filepath.Join(tmpDir, "javascript.md"): "JavaScript is a high-level, often just-in-time compiled language that conforms to the ECMAScript specification. It has dynamic typing, prototype-based object-orientation, and first-class functions.",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name          string
		query         string
		expectedName  string
		shouldContain string // Part of the expected content
	}{
		{
			name:          "Golang query",
			query:         "How do I use goroutines in Go for concurrent programming?",
			expectedName:  "Golang",
			shouldContain: "goroutines for concurrency",
		},
		{
			name:          "Python query",
			query:         "I need help with Python indentation rules",
			expectedName:  "Python",
			shouldContain: "significant indentation",
		},
		{
			name:          "JavaScript query",
			query:         "How does prototype inheritance work in JavaScript?",
			expectedName:  "Javascript",
			shouldContain: "prototype-based object-orientation",
		},
		{
			name:          "Ambiguous query should match something",
			query:         "programming language",
			expectedName:  "", // We don't care which one it matches
			shouldContain: "", // We just want to ensure it matches one of them
		},
		{
			name:          "Irrelevant query",
			query:         "How to bake a chocolate cake?",
			expectedName:  "", // Should still match something based on words like "how"
			shouldContain: "", // But we don't care what it matches
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, err := ClassifyIntent(tc.query, tmpDir, "")
			require.NoError(t, err)

			if tc.expectedName != "" {
				assert.Equal(t, tc.expectedName, match.Name)
			} else {
				assert.NotEmpty(t, match.Name)
			}

			if tc.shouldContain != "" {
				assert.Contains(t, match.Content, tc.shouldContain)
			}

			// Verify score is between 0 and 1
			assert.GreaterOrEqual(t, match.Score, 0.0)
			assert.LessOrEqual(t, match.Score, 1.0)
		})
	}
}

func TestClassifyIntentWithThreshold(t *testing.T) {
	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "threshold-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test strategy files
	testFiles := map[string]string{
		filepath.Join(tmpDir, "specific_topic.md"): "This is a very specific topic about quantum computing and neural networks",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name          string
		query         string
		threshold     float64
		expectDefault bool
	}{
		{
			name:          "Matching query above threshold",
			query:         "Tell me about quantum computing",
			threshold:     0.1,
			expectDefault: false,
		},
		{
			name:          "Matching query below threshold",
			query:         "Tell me about quantum computing",
			threshold:     0.8,
			expectDefault: true,
		},
		{
			name:          "Non-matching query",
			query:         "How to make pasta",
			threshold:     0.1,
			expectDefault: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, err := ClassifyIntentWithThreshold(tc.query, tmpDir, "", tc.threshold)
			require.NoError(t, err)

			if tc.expectDefault {
				assert.Equal(t, "Default Strategy", match.Name)
				assert.Equal(t, DefaultStrategy, match.Content)
			} else {
				assert.Equal(t, "Specific Topic", match.Name)
				assert.Contains(t, match.Content, "quantum computing")
			}
		})
	}
}

func TestGetTopNMatches(t *testing.T) {
	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "top-n-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test strategy files
	testFiles := map[string]string{
		filepath.Join(tmpDir, "database.md"): "Databases store structured data. SQL is a query language for relational databases like MySQL and PostgreSQL.",
		filepath.Join(tmpDir, "frontend.md"): "Frontend development involves HTML, CSS, and JavaScript to create user interfaces for websites and applications.",
		filepath.Join(tmpDir, "backend.md"):  "Backend development involves server-side logic, databases, and APIs. Languages like Go, Java, and Python are commonly used.",
		filepath.Join(tmpDir, "devops.md"):   "DevOps integrates development and operations. It involves practices like CI/CD, containerization with Docker, and orchestration with Kubernetes.",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test getting top matches
	query := "How do I create a REST API with Go and connect it to a PostgreSQL database?"
	matches, err := GetTopNMatches(query, tmpDir, "", 3)
	require.NoError(t, err)

	// Should have 3 matches
	assert.Equal(t, 3, len(matches))

	// First match should be backend, database, or devops since the query mentions REST API, Go, and PostgreSQL
	assert.Contains(t, []string{"Backend", "Database", "Devops"}, matches[0].Name)

	// Scores should be in descending order
	assert.GreaterOrEqual(t, matches[0].Score, matches[1].Score)
	assert.GreaterOrEqual(t, matches[1].Score, matches[2].Score)

	// Test with N greater than available strategies
	allMatches, err := GetTopNMatches(query, tmpDir, "", 10)
	require.NoError(t, err)
	assert.Equal(t, 4, len(allMatches))

	// Test with N=1
	topMatch, err := GetTopNMatches(query, tmpDir, "", 1)
	require.NoError(t, err)
	assert.Equal(t, 1, len(topMatch))
}
