package llm

import (
	"testing"
)

// TestParser_RealLLMFormat 测试真实 LLM 返回的字段格式
// 真实 LLM 常返回 file, line, message 而非 file_path, line_start, description
func TestParser_RealLLMFormat(t *testing.T) {
	// 真实的 gpt-5.1-codex-max 输出格式
	realLLMOutput := `<think>**Providing detailed code review summary and JSON schema**

I'm preparing a JSON summary with an 88 score reflecting minor warnings.</think>

{
  "summary": "Added gfcli config entries for contract_po and purchase tables and generated DAO for contract_po.",
  "score": 88,
  "suggestions": [
    {
      "file": "hack/config.yaml",
      "line": 9,
      "severity": "warning",
      "message": "Database credentials are committed in plaintext (new entry duplicates the secret), which is a security risk.",
      "suggestion": "Move credentials to environment variables or a local-only config template and reference them via placeholders so secrets are not stored in VCS."
    },
    {
      "file": "hack/config.yaml",
      "line": 12,
      "severity": "warning",
      "message": "New gfcli generator entry for logic/srm/internal/dao (tables: purchase) is added, but no generated DAO code for this target is in the diff.",
      "suggestion": "Run the gfcli generator for the new entry and commit the generated DAO files (or confirm generation is handled elsewhere) to avoid missing code at build time."
    }
  ]
}`

	// 使用 parser 解析
	reviewResp, err := parseReviewResponse(realLLMOutput)
	if err != nil {
		t.Fatalf("parseReviewResponse() error = %v", err)
	}

	// 验证基本字段
	t.Run("BasicFields", func(t *testing.T) {
		if reviewResp.Summary == "" {
			t.Error("Summary should not be empty")
		}
		if reviewResp.Score != 88 {
			t.Errorf("Score = %d, want 88", reviewResp.Score)
		}
		if len(reviewResp.Suggestions) != 2 {
			t.Errorf("len(Suggestions) = %d, want 2", len(reviewResp.Suggestions))
		}
	})

	// 验证字段映射：file → FilePath
	t.Run("FieldMapping_File", func(t *testing.T) {
		if len(reviewResp.Suggestions) == 0 {
			t.Fatal("No suggestions found")
		}
		
		firstSuggestion := reviewResp.Suggestions[0]
		if firstSuggestion.FilePath != "hack/config.yaml" {
			t.Errorf("FilePath = %q, want 'hack/config.yaml'", firstSuggestion.FilePath)
		}
	})

	// 验证字段映射：line → LineStart
	t.Run("FieldMapping_Line", func(t *testing.T) {
		if len(reviewResp.Suggestions) == 0 {
			t.Fatal("No suggestions found")
		}
		
		firstSuggestion := reviewResp.Suggestions[0]
		if firstSuggestion.LineStart != 9 {
			t.Errorf("LineStart = %d, want 9", firstSuggestion.LineStart)
		}
		
		secondSuggestion := reviewResp.Suggestions[1]
		if secondSuggestion.LineStart != 12 {
			t.Errorf("LineStart = %d, want 12", secondSuggestion.LineStart)
		}
	})

	// 验证字段映射：message → Description
	t.Run("FieldMapping_Message", func(t *testing.T) {
		if len(reviewResp.Suggestions) == 0 {
			t.Fatal("No suggestions found")
		}
		
		firstSuggestion := reviewResp.Suggestions[0]
		if firstSuggestion.Description == "" {
			t.Error("Description should not be empty")
		}
		if firstSuggestion.Description != "Database credentials are committed in plaintext (new entry duplicates the secret), which is a security risk." {
			t.Errorf("Description = %q, want full message", firstSuggestion.Description)
		}
	})

	// 验证 suggestion 字段保持不变
	t.Run("FieldMapping_Suggestion", func(t *testing.T) {
		if len(reviewResp.Suggestions) == 0 {
			t.Fatal("No suggestions found")
		}
		
		firstSuggestion := reviewResp.Suggestions[0]
		if firstSuggestion.Suggestion == "" {
			t.Error("Suggestion should not be empty")
		}
		if firstSuggestion.Suggestion != "Move credentials to environment variables or a local-only config template and reference them via placeholders so secrets are not stored in VCS." {
			t.Errorf("Suggestion = %q, want full suggestion", firstSuggestion.Suggestion)
		}
	})

	// 验证 severity 标准化
	t.Run("Severity_Normalization", func(t *testing.T) {
		if len(reviewResp.Suggestions) == 0 {
			t.Fatal("No suggestions found")
		}
		
		// "warning" 应该被标准化为 "medium"
		for i, sug := range reviewResp.Suggestions {
			if sug.Severity != "medium" {
				t.Errorf("Suggestions[%d].Severity = %q, want 'medium' (normalized from 'warning')", i, sug.Severity)
			}
		}
	})
}

// TestParser_MixedFieldFormat 测试混合字段格式的兼容性
// 某些 LLM 可能同时返回 file 和 file_path
func TestParser_MixedFieldFormat(t *testing.T) {
	mixedFormatOutput := `{
  "summary": "Test mixed format",
  "score": 75,
  "suggestions": [
    {
      "file_path": "src/main.go",
      "line_start": 10,
      "description": "Use file_path format",
      "suggestion": "Fix this"
    },
    {
      "file": "src/util.go",
      "line": 20,
      "message": "Use file format",
      "suggestion": "Fix that"
    }
  ]
}`

	reviewResp, err := parseReviewResponse(mixedFormatOutput)
	if err != nil {
		t.Fatalf("parseReviewResponse() error = %v", err)
	}

	if len(reviewResp.Suggestions) != 2 {
		t.Fatalf("len(Suggestions) = %d, want 2", len(reviewResp.Suggestions))
	}

	// 第一个使用 file_path/line_start/description
	t.Run("PreferSpecificFields", func(t *testing.T) {
		sug1 := reviewResp.Suggestions[0]
		if sug1.FilePath != "src/main.go" {
			t.Errorf("Suggestions[0].FilePath = %q, want 'src/main.go'", sug1.FilePath)
		}
		if sug1.LineStart != 10 {
			t.Errorf("Suggestions[0].LineStart = %d, want 10", sug1.LineStart)
		}
		if sug1.Description != "Use file_path format" {
			t.Errorf("Suggestions[0].Description = %q, want 'Use file_path format'", sug1.Description)
		}
	})

	// 第二个使用 file/line/message（回退字段）
	t.Run("FallbackToGenericFields", func(t *testing.T) {
		sug2 := reviewResp.Suggestions[1]
		if sug2.FilePath != "src/util.go" {
			t.Errorf("Suggestions[1].FilePath = %q, want 'src/util.go'", sug2.FilePath)
		}
		if sug2.LineStart != 20 {
			t.Errorf("Suggestions[1].LineStart = %d, want 20", sug2.LineStart)
		}
		if sug2.Description != "Use file format" {
			t.Errorf("Suggestions[1].Description = %q, want 'Use file format'", sug2.Description)
		}
	})
}
