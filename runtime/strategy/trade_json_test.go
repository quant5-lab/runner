package strategy

import (
	"encoding/json"
	"testing"
	"time"
)

/* TestTradeJSONSerialization verifies entryComment and exitComment JSON marshaling */
func TestTradeJSONSerialization(t *testing.T) {
	trade := Trade{
		EntryID:      "long1",
		Direction:    Long,
		Size:         1.0,
		EntryPrice:   100.0,
		EntryBar:     10,
		EntryTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).Unix(),
		EntryComment: "Buy signal",
		ExitPrice:    110.0,
		ExitBar:      20,
		ExitTime:     time.Date(2024, 1, 2, 15, 30, 0, 0, time.UTC).Unix(),
		ExitComment:  "Take profit",
		Profit:       10.0,
	}

	/* Serialize to JSON */
	data, err := json.Marshal(trade)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	/* Verify entryComment and exitComment in JSON output */
	jsonStr := string(data)
	if !containsSubstring(jsonStr, `"entryComment":"Buy signal"`) {
		t.Errorf("Expected entryComment in JSON, got: %s", jsonStr)
	}
	if !containsSubstring(jsonStr, `"exitComment":"Take profit"`) {
		t.Errorf("Expected exitComment in JSON, got: %s", jsonStr)
	}

	/* Deserialize and verify */
	var decoded Trade
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded.EntryComment != "Buy signal" {
		t.Errorf("Expected EntryComment 'Buy signal', got %q", decoded.EntryComment)
	}
	if decoded.ExitComment != "Take profit" {
		t.Errorf("Expected ExitComment 'Take profit', got %q", decoded.ExitComment)
	}
}

/* TestTradeJSONSerializationEmptyComments verifies empty string comment handling */
func TestTradeJSONSerializationEmptyComments(t *testing.T) {
	trade := Trade{
		EntryID:      "long2",
		Direction:    Long,
		Size:         1.0,
		EntryPrice:   100.0,
		EntryBar:     10,
		EntryTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).Unix(),
		EntryComment: "",
		ExitPrice:    105.0,
		ExitBar:      15,
		ExitTime:     time.Date(2024, 1, 2, 15, 30, 0, 0, time.UTC).Unix(),
		ExitComment:  "",
		Profit:       5.0,
	}

	/* Serialize to JSON */
	data, err := json.Marshal(trade)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	/* Verify empty strings serialized correctly */
	jsonStr := string(data)
	if !containsSubstring(jsonStr, `"entryComment":""`) {
		t.Errorf("Expected empty entryComment in JSON, got: %s", jsonStr)
	}
	if !containsSubstring(jsonStr, `"exitComment":""`) {
		t.Errorf("Expected empty exitComment in JSON, got: %s", jsonStr)
	}

	/* Deserialize and verify */
	var decoded Trade
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded.EntryComment != "" {
		t.Errorf("Expected empty EntryComment, got %q", decoded.EntryComment)
	}
	if decoded.ExitComment != "" {
		t.Errorf("Expected empty ExitComment, got %q", decoded.ExitComment)
	}
}

/* TestTradeJSONSerializationOpenTrade verifies open trade comment handling */
func TestTradeJSONSerializationOpenTrade(t *testing.T) {
	trade := Trade{
		EntryID:      "long3",
		Direction:    Long,
		Size:         1.0,
		EntryPrice:   100.0,
		EntryBar:     10,
		EntryTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).Unix(),
		EntryComment: "Trend following entry",
		ExitComment:  "", // Open trades have no exit comment yet
	}

	/* Serialize to JSON */
	data, err := json.Marshal(trade)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	/* Verify entryComment present, exitComment empty for open trade */
	jsonStr := string(data)
	if !containsSubstring(jsonStr, `"entryComment":"Trend following entry"`) {
		t.Errorf("Expected entryComment in JSON, got: %s", jsonStr)
	}
	if !containsSubstring(jsonStr, `"exitComment":""`) {
		t.Errorf("Expected empty exitComment for open trade, got: %s", jsonStr)
	}

	/* Deserialize and verify */
	var decoded Trade
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded.EntryComment != "Trend following entry" {
		t.Errorf("Expected EntryComment 'Trend following entry', got %q", decoded.EntryComment)
	}
	if decoded.ExitComment != "" {
		t.Errorf("Expected empty ExitComment for open trade, got %q", decoded.ExitComment)
	}
}

/* TestTradeJSONSerializationSpecialCharacters verifies special character escaping */
func TestTradeJSONSerializationSpecialCharacters(t *testing.T) {
	trade := Trade{
		EntryID:      "long4",
		Direction:    Long,
		Size:         1.0,
		EntryPrice:   100.0,
		EntryBar:     10,
		EntryTime:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).Unix(),
		EntryComment: `Signal: "buy"`,
		ExitPrice:    110.0,
		ExitBar:      20,
		ExitTime:     time.Date(2024, 1, 2, 15, 30, 0, 0, time.UTC).Unix(),
		ExitComment:  "Exit: level\nreached",
		Profit:       10.0,
	}

	/* Serialize to JSON */
	data, err := json.Marshal(trade)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	/* Deserialize and verify special characters preserved */
	var decoded Trade
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if decoded.EntryComment != `Signal: "buy"` {
		t.Errorf("Expected quotes preserved, got %q", decoded.EntryComment)
	}
	if decoded.ExitComment != "Exit: level\nreached" {
		t.Errorf("Expected newline preserved, got %q", decoded.ExitComment)
	}
}

/* Helper function to check substring presence (case-sensitive) */
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
