package api

import "math"

// CosineSimilarity calculates the cosine similarity between two embeddings.
// This is now the single implementation of this function in the codebase.
func CosineSimilarity(vec1, vec2 []float32) float32 {
	if len(vec1) == 0 || len(vec2) == 0 || len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, norm1, norm2 float32

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	// Avoid division by zero
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	// Compute cosine similarity: dot(vec1,vec2)/(||vec1||*||vec2||)
	return dotProduct / (float32(math.Sqrt(float64(norm1))) * float32(math.Sqrt(float64(norm2))))
}
