package intent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClassifyIntentWithEmbeddings tests embedding-based classification
// This test will be skipped if OPENAI_API_KEY is not set
func TestClassifyIntentWithEmbeddings(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "embedding-classify-test-*")
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
			name:          "Semantic query without exact keyword match",
			query:         "What's the best language for web development?",
			expectedName:  "Javascript", // Embeddings should semantically match this to JavaScript
			shouldContain: "JavaScript",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, err := ClassifyIntentWithEmbeddings(tc.query, tmpDir, "")
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

// TestClassifyIntentWithEmbeddingsThreshold tests threshold-based embedding classification
// This test will be skipped if OPENAI_API_KEY is not set
func TestClassifyIntentWithEmbeddingsThreshold(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "embedding-threshold-test-*")
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
			threshold:     0.5,
			expectDefault: false,
		},
		{
			name:          "Matching query below threshold",
			query:         "Tell me about quantum computing",
			threshold:     0.9,
			expectDefault: true,
		},
		{
			name:          "Non-matching query",
			query:         "How to make pasta",
			threshold:     0.9,
			expectDefault: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, err := ClassifyIntentWithEmbeddingsThreshold(tc.query, tmpDir, "", tc.threshold)
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

// TestGetTopNMatchesWithEmbeddings tests getting top N matches with embeddings
// This test will be skipped if OPENAI_API_KEY is not set
func TestGetTopNMatchesWithEmbeddings(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "embedding-top-n-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test strategy files with more semantic content
	testFiles := map[string]string{
		filepath.Join(tmpDir, "database.md"): "Databases store structured data. SQL is a query language for relational databases like MySQL and PostgreSQL. Database optimization includes indexing, query tuning, and schema design.",
		filepath.Join(tmpDir, "frontend.md"): "Frontend development involves HTML, CSS, and JavaScript to create user interfaces for websites and applications. React, Vue, and Angular are popular frontend frameworks.",
		filepath.Join(tmpDir, "backend.md"):  "Backend development involves server-side logic, databases, and APIs. Languages like Go, Java, and Python are commonly used. RESTful APIs are a standard approach for web services.",
		filepath.Join(tmpDir, "devops.md"):   "DevOps integrates development and operations. It involves practices like CI/CD, containerization with Docker, and orchestration with Kubernetes. Infrastructure as code is a key concept.",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test getting top matches
	query := "How do I create a REST API with Go and connect it to a PostgreSQL database?"
	matches, err := GetTopNMatchesWithEmbeddings(query, tmpDir, "", 3)
	require.NoError(t, err)

	// Should have 3 matches
	assert.Equal(t, 3, len(matches))

	// First match should be backend or database since the query mentions both
	assert.Contains(t, []string{"Backend", "Database"}, matches[0].Name)

	// Scores should be in descending order
	assert.GreaterOrEqual(t, matches[0].Score, matches[1].Score)
	assert.GreaterOrEqual(t, matches[1].Score, matches[2].Score)

	// Test with N greater than available strategies
	allMatches, err := GetTopNMatchesWithEmbeddings(query, tmpDir, "", 10)
	require.NoError(t, err)
	assert.Equal(t, 4, len(allMatches))

	// Test with N=1
	topMatch, err := GetTopNMatchesWithEmbeddings(query, tmpDir, "", 1)
	require.NoError(t, err)
	assert.Equal(t, 1, len(topMatch))

	// Semantic query test
	semanticQuery := "What's the best way to build a web API?"
	semanticMatches, err := GetTopNMatchesWithEmbeddings(semanticQuery, tmpDir, "", 2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(semanticMatches))

	// Should match backend even though "web API" isn't explicitly mentioned
	assert.Contains(t, semanticMatches[0].Name, "Backend")
}

// TestEmbeddingPerformanceComparison compares token-based vs embedding-based classification
// This is an optional test that demonstrates the improved accuracy of embeddings
// This test will be skipped if OPENAI_API_KEY is not set
func TestEmbeddingPerformanceComparison(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Create a temporary directory with test strategy files
	tmpDir, err := os.MkdirTemp("", "embedding-comparison-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test strategy files with challenging semantic content
	// These examples are designed to show where embeddings outperform token matching
	testFiles := map[string]string{
		filepath.Join(tmpDir, "database.md"):         "Database systems manage structured information. They provide mechanisms for storing, retrieving, and managing data.",
		filepath.Join(tmpDir, "machine_learning.md"): "Machine learning enables computers to learn patterns from data without explicit programming. Neural networks are a popular approach.",
	}

	for path, content := range testFiles {
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Semantic query without exact keyword matches
	query := "How can I train an AI model on my dataset?"

	// Get results from both methods
	tokenMatch, err := ClassifyIntent(query, tmpDir, "")
	require.NoError(t, err)

	embeddingMatch, err := ClassifyIntentWithEmbeddings(query, tmpDir, "")
	require.NoError(t, err)

	// Token-based match might choose randomly or incorrectly
	// but embedding-based should correctly identify machine learning
	assert.Equal(t, "Machine Learning", embeddingMatch.Name)

	// Log results for debugging
	t.Logf("Token-based match: %s (score: %.4f)", tokenMatch.Name, tokenMatch.Score)
	t.Logf("Embedding-based match: %s (score: %.4f)", embeddingMatch.Name, embeddingMatch.Score)
}
