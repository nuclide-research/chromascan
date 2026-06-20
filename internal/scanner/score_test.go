package scanner

import "testing"

func TestScoreCollection(t *testing.T) {
	cases := []struct {
		name           string
		authOpen       bool
		piiSignals     []string
		recordCount    int64
		writeConfirmed bool
		wantMin        float64
		wantLabel      string
	}{
		{
			name:        "auth closed, no pii",
			authOpen:    false,
			recordCount: 100,
			wantMin:     0,
			wantLabel:   "LOW",
		},
		{
			name:        "auth open, no pii, small corpus",
			authOpen:    true,
			recordCount: 100,
			wantMin:     4.0,
			wantLabel:   "MEDIUM",
		},
		{
			name:        "auth open, pii, large corpus",
			authOpen:    true,
			piiSignals:  []string{"email"},
			recordCount: 15000,
			wantMin:     7.5,
			wantLabel:   "HIGH",
		},
		{
			name:           "auth open, pii, write confirmed",
			authOpen:       true,
			piiSignals:     []string{"email", "medical"},
			recordCount:    15000,
			writeConfirmed: true,
			wantMin:        9.0,
			wantLabel:      "CRITICAL",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			score := scoreCollection(tc.authOpen, tc.piiSignals, tc.recordCount, tc.writeConfirmed)
			if score < tc.wantMin {
				t.Errorf("score %.1f < wantMin %.1f", score, tc.wantMin)
			}
			label := severityLabel(score)
			if label != tc.wantLabel {
				t.Errorf("label %q != want %q (score=%.1f)", label, tc.wantLabel, score)
			}
		})
	}
}

func TestSeverityLabel(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{0.0, "LOW"},
		{3.9, "LOW"},
		{4.0, "MEDIUM"},
		{5.9, "MEDIUM"},
		{6.0, "HIGH"},
		{7.9, "HIGH"},
		{8.0, "CRITICAL"},
		{9.5, "CRITICAL"},
	}
	for _, tc := range cases {
		got := severityLabel(tc.score)
		if got != tc.want {
			t.Errorf("severityLabel(%.1f) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
