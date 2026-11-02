package cli_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"corridor/internal/cli"
)

func TestOutreach_Default(t *testing.T) {
	outDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	ctx := context.Background()

	err := cli.Run(ctx, []string{"outreach", "--out=" + outDir}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("outreach default failed: %v, stderr: %s", err, stderr.String())
	}

	outLog := stdout.String()
	if !strings.Contains(outLog, "Federal outreach package generated successfully") {
		t.Errorf("expected success message in stdout, got:\n%s", outLog)
	}

	// 1. Verify 24 .eml files in emails/
	emailsDir := filepath.Join(outDir, "emails")
	emailEntries, err := os.ReadDir(emailsDir)
	if err != nil {
		t.Fatalf("failed to read emails dir %s: %v", emailsDir, err)
	}

	if len(emailEntries) != 24 {
		t.Errorf("expected 24 email draft files, got %d", len(emailEntries))
	}

	expectedParks := []string{
		"yellowstone", "grand_teton", "glacier", "rocky_mountain",
		"yosemite", "grand_canyon", "olympic", "shenandoah",
	}
	expectedPersonas := []string{"biologist", "ranger", "superintendent"}

	for _, park := range expectedParks {
		for _, persona := range expectedPersonas {
			emlFile := filepath.Join(emailsDir, park+"_"+persona+".eml")
			rawBytes, err := os.ReadFile(emlFile)
			if err != nil {
				t.Errorf("missing expected email file %s: %v", emlFile, err)
				continue
			}

			// Parse MIME email
			msg, err := mail.ReadMessage(bytes.NewReader(rawBytes))
			if err != nil {
				t.Errorf("invalid RFC-822 MIME message %s: %v", emlFile, err)
				continue
			}

			if msg.Header.Get("From") != "Corridor Wildlife Project <adoption@corridorwildlife.org>" {
				t.Errorf("%s: unexpected From header: %q", emlFile, msg.Header.Get("From"))
			}

			toHeader := msg.Header.Get("To")
			if !strings.Contains(toHeader, "@nps.gov") {
				t.Errorf("%s: To header must contain an authoritative @nps.gov address: %q", emlFile, toHeader)
			}

			if msg.Header.Get("Subject") == "" {
				t.Errorf("%s: missing Subject header", emlFile)
			}

			dateHeader := msg.Header.Get("Date")
			if _, err := time.Parse(time.RFC1123Z, dateHeader); err != nil {
				t.Errorf("%s: invalid RFC1123Z Date header %q: %v", emlFile, dateHeader, err)
			}

			msgID := msg.Header.Get("Message-ID")
			if !strings.HasPrefix(msgID, "<") || !strings.HasSuffix(msgID, ">") || !strings.Contains(msgID, "@corridorwildlife.org") {
				t.Errorf("%s: invalid Message-ID header: %q", emlFile, msgID)
			}

			if msg.Header.Get("MIME-Version") != "1.0" {
				t.Errorf("%s: expected MIME-Version 1.0, got: %q", emlFile, msg.Header.Get("MIME-Version"))
			}

			if msg.Header.Get("X-Unsent") != "1" {
				t.Errorf("%s: expected X-Unsent 1 header for desktop mail drafts, got: %q", emlFile, msg.Header.Get("X-Unsent"))
			}

			bodyBytes, err := io.ReadAll(msg.Body)
			if err != nil {
				t.Errorf("%s: failed to read body: %v", emlFile, err)
				continue
			}
			body := string(bodyBytes)

			// Verify USDOT crash valuations are cited
			if !strings.Contains(body, "$14,000") || !strings.Contains(body, "$45,000") || !strings.Contains(body, "$83,000") {
				t.Errorf("%s: body missing USDOT crash valuations: %s", emlFile, body)
			}

			// Verify park-specific highway corridor is cited
			switch park {
			case "yellowstone":
				if !strings.Contains(body, "US-191") && !strings.Contains(body, "US-89") {
					t.Errorf("%s: body missing Yellowstone corridors (US-191 / US-89)", emlFile)
				}
			case "grand_teton":
				if !strings.Contains(body, "US-89") {
					t.Errorf("%s: body missing Grand Teton corridors (US-89)", emlFile)
				}
			case "glacier":
				if !strings.Contains(body, "US-2") {
					t.Errorf("%s: body missing Glacier corridors (US-2)", emlFile)
				}
			case "rocky_mountain":
				if !strings.Contains(body, "US-34") {
					t.Errorf("%s: body missing Rocky Mountain corridors (US-34)", emlFile)
				}
			case "yosemite":
				if !strings.Contains(body, "CA-120") {
					t.Errorf("%s: body missing Yosemite corridors (CA-120)", emlFile)
				}
			case "grand_canyon":
				if !strings.Contains(body, "AZ-64") {
					t.Errorf("%s: body missing Grand Canyon corridors (AZ-64)", emlFile)
				}
			case "olympic":
				if !strings.Contains(body, "US-101") {
					t.Errorf("%s: body missing Olympic corridors (US-101)", emlFile)
				}
			case "shenandoah":
				if !strings.Contains(body, "Skyline Drive") {
					t.Errorf("%s: body missing Shenandoah corridors (Skyline Drive)", emlFile)
				}
			}
		}
	}

	// 2. Verify 8 Park Briefing Pack Markdown files in briefs/
	briefsDir := filepath.Join(outDir, "briefs")
	briefEntries, err := os.ReadDir(briefsDir)
	if err != nil {
		t.Fatalf("failed to read briefs dir %s: %v", briefsDir, err)
	}

	if len(briefEntries) != 8 {
		t.Errorf("expected 8 park briefing files, got %d", len(briefEntries))
	}

	for _, park := range expectedParks {
		briefFile := filepath.Join(briefsDir, park+"_brief.md")
		data, err := os.ReadFile(briefFile)
		if err != nil {
			t.Errorf("missing expected brief file %s: %v", briefFile, err)
			continue
		}

		content := string(data)
		if !strings.Contains(content, "23 U.S.C. § 171") {
			t.Errorf("%s: missing 23 U.S.C. § 171 statutory reference", briefFile)
		}
		if !strings.Contains(content, "miles") {
			t.Errorf("%s: missing highway mileage detail", briefFile)
		}
		if !strings.Contains(content, "Focal Species") && !strings.Contains(content, "Focal Wildlife Species") {
			t.Errorf("%s: missing focal species section", briefFile)
		}
		if !strings.Contains(content, "Bottleneck") && !strings.Contains(content, "Hotspots") {
			t.Errorf("%s: missing migration bottleneck section", briefFile)
		}
		if !strings.Contains(content, "$14,000") || !strings.Contains(content, "$45,000") || !strings.Contains(content, "$83,000") {
			t.Errorf("%s: missing USDOT crash cost standards", briefFile)
		}
		if !strings.Contains(content, "86%") {
			t.Errorf("%s: missing peer-reviewed 86%% mitigation efficacy figure", briefFile)
		}
	}

	// 3. Verify CRM Outreach Ledger CSV
	csvFile := filepath.Join(outDir, "registry.csv")
	f, err := os.Open(csvFile)
	if err != nil {
		t.Fatalf("failed to open CRM registry CSV %s: %v", csvFile, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV from %s: %v", csvFile, err)
	}

	expectedHeaders := []string{
		"Park", "Unit Code", "State", "Persona", "Recipient Name",
		"Recipient Email", "Highway Corridor", "Statutory Focus",
		"Contact Status", "Last Generated",
	}

	if len(rows) != 25 { // 1 header + 24 records
		t.Fatalf("expected 25 rows in registry.csv (1 header + 24 records), got %d", len(rows))
	}

	for i, col := range expectedHeaders {
		if rows[0][i] != col {
			t.Errorf("header column %d: expected %q, got %q", i, col, rows[0][i])
		}
	}

	for rowIdx, row := range rows[1:] {
		park := row[0]
		unitCode := row[1]
		state := row[2]
		persona := row[3]
		name := row[4]
		email := row[5]
		corridor := row[6]
		statutory := row[7]
		status := row[8]
		lastGen := row[9]

		if park == "" || unitCode == "" || len(state) != 2 || persona == "" || name == "" {
			t.Errorf("row %d has empty core fields: %v", rowIdx+1, row)
		}
		if !strings.Contains(email, "@nps.gov") {
			t.Errorf("row %d: recipient email %q must be an @nps.gov address", rowIdx+1, email)
		}
		if corridor == "" || statutory == "" || status == "" || lastGen == "" {
			t.Errorf("row %d has empty auxiliary fields: %v", rowIdx+1, row)
		}
	}
}

func TestOutreach_ParkFilter(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		filter       string
		expectedSlug string
		expectedCode string
	}{
		{"Yellowstone", "yellowstone", "YELL"},
		{"yellowstone", "yellowstone", "YELL"},
		{"YELL", "yellowstone", "YELL"},
		{"Grand Teton", "grand_teton", "GRTE"},
		{"grand-teton", "grand_teton", "GRTE"},
		{"Glacier", "glacier", "GLAC"},
		{"Rocky Mountain", "rocky_mountain", "ROMO"},
		{"Yosemite", "yosemite", "YOSE"},
		{"Grand Canyon", "grand_canyon", "GRCA"},
		{"Olympic", "olympic", "OLYM"},
		{"Shenandoah", "shenandoah", "SHEN"},
	}

	for _, tc := range testCases {
		t.Run(tc.filter, func(t *testing.T) {
			outDir := t.TempDir()
			var stdout, stderr bytes.Buffer

			err := cli.Run(ctx, []string{"outreach", "--park=" + tc.filter, "--out=" + outDir}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("outreach with park filter %q failed: %v", tc.filter, err)
			}

			// Expect 3 emails (one for each persona)
			emails, err := os.ReadDir(filepath.Join(outDir, "emails"))
			if err != nil {
				t.Fatalf("failed to read emails dir: %v", err)
			}
			if len(emails) != 3 {
				t.Errorf("expected 3 emails for park %s, got %d", tc.filter, len(emails))
			}

			// Expect 1 brief
			briefs, err := os.ReadDir(filepath.Join(outDir, "briefs"))
			if err != nil {
				t.Fatalf("failed to read briefs dir: %v", err)
			}
			if len(briefs) != 1 {
				t.Errorf("expected 1 brief for park %s, got %d", tc.filter, len(briefs))
			}
			expectedBrief := tc.expectedSlug + "_brief.md"
			if briefs[0].Name() != expectedBrief {
				t.Errorf("expected brief name %s, got %s", expectedBrief, briefs[0].Name())
			}

			// Expect 4 rows in CSV (header + 3 personas)
			csvData, err := os.ReadFile(filepath.Join(outDir, "registry.csv"))
			if err != nil {
				t.Fatalf("failed to read registry CSV: %v", err)
			}
			reader := csv.NewReader(bytes.NewReader(csvData))
			rows, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("failed to parse CSV: %v", err)
			}
			if len(rows) != 4 {
				t.Errorf("expected 4 rows in CSV for park %s, got %d", tc.filter, len(rows))
			}
			for _, r := range rows[1:] {
				if r[1] != tc.expectedCode {
					t.Errorf("expected unit code %s, got %s", tc.expectedCode, r[1])
				}
			}
		})
	}
}

func TestOutreach_PersonaFilter(t *testing.T) {
	ctx := context.Background()

	personas := []string{"biologist", "ranger", "superintendent"}
	for _, persona := range personas {
		t.Run(persona, func(t *testing.T) {
			outDir := t.TempDir()
			var stdout, stderr bytes.Buffer

			err := cli.Run(ctx, []string{"outreach", "--persona=" + persona, "--out=" + outDir}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("outreach with persona filter %q failed: %v", persona, err)
			}

			emails, err := os.ReadDir(filepath.Join(outDir, "emails"))
			if err != nil {
				t.Fatalf("failed to read emails dir: %v", err)
			}
			if len(emails) != 8 { // 8 parks * 1 persona
				t.Errorf("expected 8 emails for persona %s, got %d", persona, len(emails))
			}

			for _, e := range emails {
				if !strings.HasSuffix(e.Name(), "_"+persona+".eml") {
					t.Errorf("expected email file ending with _%s.eml, got %s", persona, e.Name())
				}
			}

			// CSV should contain 1 header + 8 rows
			csvData, err := os.ReadFile(filepath.Join(outDir, "registry.csv"))
			if err != nil {
				t.Fatalf("failed to read registry CSV: %v", err)
			}
			reader := csv.NewReader(bytes.NewReader(csvData))
			rows, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("failed to parse CSV: %v", err)
			}
			if len(rows) != 9 {
				t.Errorf("expected 9 rows in CSV, got %d", len(rows))
			}
			for _, r := range rows[1:] {
				if r[3] != persona {
					t.Errorf("expected persona column %s, got %s", persona, r[3])
				}
			}
		})
	}
}

func TestOutreach_FormatFilter(t *testing.T) {
	ctx := context.Background()

	// 1. Format: eml only
	t.Run("eml_only", func(t *testing.T) {
		outDir := t.TempDir()
		var stdout, stderr bytes.Buffer
		err := cli.Run(ctx, []string{"outreach", "--format=eml", "--out=" + outDir}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("format=eml failed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(outDir, "emails")); err != nil {
			t.Errorf("expected emails directory to exist")
		}
		if _, err := os.Stat(filepath.Join(outDir, "briefs")); !os.IsNotExist(err) {
			t.Errorf("briefs directory should not exist when format=eml")
		}
		if _, err := os.Stat(filepath.Join(outDir, "registry.csv")); !os.IsNotExist(err) {
			t.Errorf("registry.csv should not exist when format=eml")
		}
	})

	// 2. Format: brief only
	t.Run("brief_only", func(t *testing.T) {
		outDir := t.TempDir()
		var stdout, stderr bytes.Buffer
		err := cli.Run(ctx, []string{"outreach", "--format=brief", "--out=" + outDir}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("format=brief failed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(outDir, "briefs")); err != nil {
			t.Errorf("expected briefs directory to exist")
		}
		if _, err := os.Stat(filepath.Join(outDir, "emails")); !os.IsNotExist(err) {
			t.Errorf("emails directory should not exist when format=brief")
		}
		if _, err := os.Stat(filepath.Join(outDir, "registry.csv")); !os.IsNotExist(err) {
			t.Errorf("registry.csv should not exist when format=brief")
		}
	})

	// 3. Format: csv only
	t.Run("csv_only", func(t *testing.T) {
		outDir := t.TempDir()
		var stdout, stderr bytes.Buffer
		err := cli.Run(ctx, []string{"outreach", "--format=csv", "--out=" + outDir}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("format=csv failed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(outDir, "registry.csv")); err != nil {
			t.Errorf("expected registry.csv to exist")
		}
		if _, err := os.Stat(filepath.Join(outDir, "emails")); !os.IsNotExist(err) {
			t.Errorf("emails directory should not exist when format=csv")
		}
		if _, err := os.Stat(filepath.Join(outDir, "briefs")); !os.IsNotExist(err) {
			t.Errorf("briefs directory should not exist when format=csv")
		}
	})
}

func TestOutreach_InvalidOptions(t *testing.T) {
	ctx := context.Background()

	// Invalid park
	var stdout, stderr bytes.Buffer
	err := cli.Run(ctx, []string{"outreach", "--park=Atlantis"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for invalid park, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported park") {
		t.Errorf("expected unsupported park error message, got: %v", err)
	}

	// Invalid persona
	stdout.Reset()
	stderr.Reset()
	err = cli.Run(ctx, []string{"outreach", "--persona=astronaut"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for invalid persona, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported persona") {
		t.Errorf("expected unsupported persona error message, got: %v", err)
	}

	// Invalid format
	stdout.Reset()
	stderr.Reset()
	err = cli.Run(ctx, []string{"outreach", "--format=docx"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for invalid format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected unsupported format error message, got: %v", err)
	}
}

func TestOutreach_Help(t *testing.T) {
	ctx := context.Background()
	var stdout, stderr bytes.Buffer

	// Test -h on outreach
	err := cli.Run(ctx, []string{"outreach", "-h"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected nil error on -h, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "-park") {
		t.Errorf("expected -park flag in stderr help output, got:\n%s", stderr.String())
	}
}
