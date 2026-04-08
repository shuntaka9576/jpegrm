package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestJPEG creates a minimal JPEG file with EXIF DateTimeOriginal.
// dateTime must be in "YYYY:MM:DD HH:MM:SS" format (19 chars).
func createTestJPEG(t *testing.T, path string, dateTime string) {
	t.Helper()

	dateBytes := append([]byte(dateTime), 0x00) // 20 bytes with null terminator

	// TIFF data (little endian)
	tiff := []byte{
		// Header: "II" (little endian), magic 42, offset to IFD0 = 8
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00,
		// IFD0: 1 entry
		0x01, 0x00,
		// Entry: ExifIFD pointer (tag=0x8769, type=LONG, count=1, value=26)
		0x69, 0x87, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1A, 0x00, 0x00, 0x00,
		// Next IFD offset = 0
		0x00, 0x00, 0x00, 0x00,
		// ExifIFD at offset 26: 1 entry
		0x01, 0x00,
		// Entry: DateTimeOriginal (tag=0x9003, type=ASCII, count=20, offset=44)
		0x03, 0x90, 0x02, 0x00, 0x14, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00,
		// Next IFD offset = 0
		0x00, 0x00, 0x00, 0x00,
	}
	// Date string at offset 44
	tiff = append(tiff, dateBytes...)

	exifHeader := append([]byte("Exif"), 0x00, 0x00)
	app1Payload := append(exifHeader, tiff...)
	app1Len := uint16(len(app1Payload) + 2) // +2 for length field itself

	var jpeg []byte
	jpeg = append(jpeg, 0xFF, 0xD8)                                  // SOI
	jpeg = append(jpeg, 0xFF, 0xE1)                                  // APP1 marker
	jpeg = append(jpeg, byte(app1Len>>8), byte(app1Len&0xFF))        // APP1 length (big endian)
	jpeg = append(jpeg, app1Payload...)                               // EXIF data
	jpeg = append(jpeg, 0xFF, 0xD9)                                  // EOI

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, jpeg, 0644); err != nil {
		t.Fatal(err)
	}
}

// createTestJPEGFull creates a minimal JPEG file with EXIF DateTimeOriginal, Make, and Model.
func createTestJPEGFull(t *testing.T, path string, dateTime string, makeName string, modelName string) {
	t.Helper()

	dateBytes := append([]byte(dateTime), 0x00)
	makeBytes := append([]byte(makeName), 0x00)
	modelBytes := append([]byte(modelName), 0x00)

	// IFD0: 3 entries (Make, Model, ExifIFD pointer)
	// Each IFD entry: 12 bytes. IFD overhead: 2 (count) + 4 (next IFD offset)
	// IFD0 starts at offset 8
	// IFD0 size: 2 + 3*12 + 4 = 42 bytes → IFD0 ends at 50
	// ExifIFD starts at offset 50
	// ExifIFD size: 2 + 1*12 + 4 = 18 bytes → ExifIFD ends at 68
	// Data area starts at offset 68

	makeOffset := uint32(68)
	modelOffset := makeOffset + uint32(len(makeBytes))
	dateOffset := modelOffset + uint32(len(modelBytes))

	tiff := []byte{
		// Header: "II" (little endian), magic 42, offset to IFD0 = 8
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00,
		// IFD0: 3 entries
		0x03, 0x00,
	}

	// Entry 1: Make (tag=0x010F, type=ASCII=2)
	makeEntry := []byte{
		0x0F, 0x01, 0x02, 0x00,
		byte(len(makeBytes)), byte(len(makeBytes) >> 8), byte(len(makeBytes) >> 16), byte(len(makeBytes) >> 24),
		byte(makeOffset), byte(makeOffset >> 8), byte(makeOffset >> 16), byte(makeOffset >> 24),
	}
	tiff = append(tiff, makeEntry...)

	// Entry 2: Model (tag=0x0110, type=ASCII=2)
	modelEntry := []byte{
		0x10, 0x01, 0x02, 0x00,
		byte(len(modelBytes)), byte(len(modelBytes) >> 8), byte(len(modelBytes) >> 16), byte(len(modelBytes) >> 24),
		byte(modelOffset), byte(modelOffset >> 8), byte(modelOffset >> 16), byte(modelOffset >> 24),
	}
	tiff = append(tiff, modelEntry...)

	// Entry 3: ExifIFD pointer (tag=0x8769, type=LONG=4, count=1, value=50)
	exifEntry := []byte{
		0x69, 0x87, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x32, 0x00, 0x00, 0x00,
	}
	tiff = append(tiff, exifEntry...)

	// Next IFD offset = 0
	tiff = append(tiff, 0x00, 0x00, 0x00, 0x00)

	// ExifIFD at offset 50: 1 entry
	tiff = append(tiff, 0x01, 0x00)

	// Entry: DateTimeOriginal (tag=0x9003, type=ASCII=2, count=20)
	dateEntry := []byte{
		0x03, 0x90, 0x02, 0x00, 0x14, 0x00, 0x00, 0x00,
		byte(dateOffset), byte(dateOffset >> 8), byte(dateOffset >> 16), byte(dateOffset >> 24),
	}
	tiff = append(tiff, dateEntry...)

	// Next IFD offset = 0
	tiff = append(tiff, 0x00, 0x00, 0x00, 0x00)

	// Data area
	tiff = append(tiff, makeBytes...)
	tiff = append(tiff, modelBytes...)
	tiff = append(tiff, dateBytes...)

	exifHeader := append([]byte("Exif"), 0x00, 0x00)
	app1Payload := append(exifHeader, tiff...)
	app1Len := uint16(len(app1Payload) + 2)

	var jpeg []byte
	jpeg = append(jpeg, 0xFF, 0xD8)
	jpeg = append(jpeg, 0xFF, 0xE1)
	jpeg = append(jpeg, byte(app1Len>>8), byte(app1Len&0xFF))
	jpeg = append(jpeg, app1Payload...)
	jpeg = append(jpeg, 0xFF, 0xD9)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, jpeg, 0644); err != nil {
		t.Fatal(err)
	}
}

// createJPEGNoExif creates a minimal JPEG file without EXIF data.
func createJPEGNoExif(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{0xFF, 0xD8, 0xFF, 0xD9}, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestIsJPEG(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"photo.JPG", true},
		{"photo.JPEG", true},
		{"photo.Jpg", true},
		{"photo.JpEg", true},
		{"photo.png", false},
		{"photo.txt", false},
		{"photo", false},
		{".jpg", true},
		{"photo.jpg.bak", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isJPEG(tt.name); got != tt.want {
				t.Errorf("isJPEG(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseExif(t *testing.T) {
	t.Run("valid DateTimeOriginal", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jpg")
		createTestJPEG(t, path, "2024:03:15 14:30:00")

		got, err := parseExif(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
		if !got.date.Equal(want) {
			t.Errorf("got %v, want %v", got.date, want)
		}
	})

	t.Run("no EXIF data", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "noexif.jpg")
		createJPEGNoExif(t, path)

		_, err := parseExif(path)
		if err == nil {
			t.Error("expected error for JPEG without EXIF")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := parseExif(filepath.Join(t.TempDir(), "nonexistent.jpg"))
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("not a JPEG file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		os.WriteFile(path, []byte("hello world"), 0644)

		_, err := parseExif(path)
		if err == nil {
			t.Error("expected error for non-JPEG file")
		}
	})
}

func TestCollectFiles(t *testing.T) {
	t.Run("flat directory with mixed files", func(t *testing.T) {
		dir := t.TempDir()
		createTestJPEG(t, filepath.Join(dir, "a.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(dir, "b.JPEG"), "2024:01:01 00:00:00")
		os.WriteFile(filepath.Join(dir, "c.png"), []byte{}, 0644)
		os.WriteFile(filepath.Join(dir, "d.txt"), []byte{}, 0644)

		files, err := collectFiles(dir, false, "*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("got %d files, want 2", len(files))
		}
	})

	t.Run("recursive finds subdirectory files", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.MkdirAll(sub, 0755)
		createTestJPEG(t, filepath.Join(dir, "a.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(sub, "b.jpg"), "2024:01:01 00:00:00")

		files, err := collectFiles(dir, true, "*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("got %d files, want 2", len(files))
		}
	})

	t.Run("non-recursive ignores subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		os.MkdirAll(sub, 0755)
		createTestJPEG(t, filepath.Join(dir, "a.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(sub, "b.jpg"), "2024:01:01 00:00:00")

		files, err := collectFiles(dir, false, "*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("got %d files, want 1", len(files))
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()

		files, err := collectFiles(dir, false, "*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("got %d files, want 0", len(files))
		}
	})

	t.Run("results are sorted", func(t *testing.T) {
		dir := t.TempDir()
		createTestJPEG(t, filepath.Join(dir, "c.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(dir, "a.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(dir, "b.jpg"), "2024:01:01 00:00:00")

		files, err := collectFiles(dir, false, "*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for i := 0; i < len(files)-1; i++ {
			if files[i] >= files[i+1] {
				t.Errorf("not sorted: %s >= %s", files[i], files[i+1])
			}
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := collectFiles(filepath.Join(t.TempDir(), "nonexistent"), false, "*")
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
	})
}

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "*"},
		{"*", "*"},
		{"*.*", "*"},
		{"DSC1234", "DSC1234.*"},
		{"DSC*", "DSC*"},
		{"DSC1234.jpg", "DSC1234.jpg"},
		{"IMG_00[0-9]?", "IMG_00[0-9]?"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizePattern(tt.input); got != tt.want {
				t.Errorf("normalizePattern(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		want     bool
	}{
		{"DSC1234.jpg", "*", true},
		{"DSC1234.jpg", "DSC1234.*", true},
		{"DSC1234.jpg", "DSC*", true},
		{"IMG_0001.jpg", "DSC*", false},
		{"DSC1234.JPG", "dsc1234.*", true},
		{"dsc1234.jpg", "DSC1234.*", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.filename, tt.pattern), func(t *testing.T) {
			if got := matchesPattern(tt.filename, tt.pattern); got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.filename, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestCollectFilesWithPattern(t *testing.T) {
	t.Run("pattern filters files", func(t *testing.T) {
		dir := t.TempDir()
		createTestJPEG(t, filepath.Join(dir, "DSC1234.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(dir, "IMG_0001.jpg"), "2024:01:01 00:00:00")

		files, err := collectFiles(dir, false, "DSC*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("got %d files, want 1", len(files))
		}
	})

	t.Run("normalized bare name matches", func(t *testing.T) {
		dir := t.TempDir()
		createTestJPEG(t, filepath.Join(dir, "DSC1234.jpg"), "2024:01:01 00:00:00")
		createTestJPEG(t, filepath.Join(dir, "DSC5678.jpg"), "2024:01:01 00:00:00")

		files, err := collectFiles(dir, false, normalizePattern("DSC1234"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("got %d files, want 1", len(files))
		}
	})
}

func TestBuildPlan(t *testing.T) {
	t.Run("single file rename", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "IMG_0001.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		plan := buildPlan([]string{src}, false, "")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		if plan[0].src != src {
			t.Errorf("src = %q, want %q", plan[0].src, src)
		}
		wantDst := filepath.Join(dir, "2024_03_15_1430_00.jpg")
		if plan[0].dst != wantDst {
			t.Errorf("dst = %q, want %q", plan[0].dst, wantDst)
		}
	})

	t.Run("duplicate timestamps get sequential numbers", func(t *testing.T) {
		dir := t.TempDir()
		files := []string{
			filepath.Join(dir, "a.jpg"),
			filepath.Join(dir, "b.jpg"),
			filepath.Join(dir, "c.jpg"),
		}
		for _, f := range files {
			createTestJPEG(t, f, "2024:03:15 14:30:00")
		}

		plan := buildPlan(files, false, "")

		if len(plan) != 3 {
			t.Fatalf("got %d pairs, want 3", len(plan))
		}
		for i, p := range plan {
			wantName := fmt.Sprintf("2024_03_15_1430_%02d.jpg", i)
			if filepath.Base(p.dst) != wantName {
				t.Errorf("plan[%d] dst name = %q, want %q", i, filepath.Base(p.dst), wantName)
			}
		}
	})

	t.Run("already named correctly is skipped", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "2024_03_15_1430_00.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		plan := buildPlan([]string{src}, false, "")

		if len(plan) != 0 {
			t.Errorf("got %d pairs, want 0 (already named correctly)", len(plan))
		}
	})

	t.Run("files without EXIF are skipped", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "noexif.jpg")
		createJPEGNoExif(t, src)

		plan := buildPlan([]string{src}, false, "")

		if len(plan) != 0 {
			t.Errorf("got %d pairs, want 0", len(plan))
		}
	})

	t.Run("disk collision increments sequence number", func(t *testing.T) {
		dir := t.TempDir()
		// Pre-create a file that would collide with _00
		collision := filepath.Join(dir, "2024_03_15_1430_00.jpg")
		os.WriteFile(collision, []byte("existing"), 0644)

		src := filepath.Join(dir, "IMG_0001.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		plan := buildPlan([]string{src}, false, "")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		wantDst := filepath.Join(dir, "2024_03_15_1430_01.jpg")
		if plan[0].dst != wantDst {
			t.Errorf("dst = %q, want %q", plan[0].dst, wantDst)
		}
	})

	t.Run("different timestamps get independent numbering", func(t *testing.T) {
		dir := t.TempDir()
		files := []string{
			filepath.Join(dir, "a.jpg"),
			filepath.Join(dir, "b.jpg"),
		}
		createTestJPEG(t, files[0], "2024:03:15 10:00:00")
		createTestJPEG(t, files[1], "2024:03:15 11:00:00")

		plan := buildPlan(files, false, "")

		if len(plan) != 2 {
			t.Fatalf("got %d pairs, want 2", len(plan))
		}
		if filepath.Base(plan[0].dst) != "2024_03_15_1000_00.jpg" {
			t.Errorf("plan[0] dst = %q, want 2024_03_15_1000_00.jpg", filepath.Base(plan[0].dst))
		}
		if filepath.Base(plan[1].dst) != "2024_03_15_1100_00.jpg" {
			t.Errorf("plan[1] dst = %q, want 2024_03_15_1100_00.jpg", filepath.Base(plan[1].dst))
		}
	})

	t.Run("mixed valid and invalid files", func(t *testing.T) {
		dir := t.TempDir()
		valid := filepath.Join(dir, "IMG_0001.jpg")
		invalid := filepath.Join(dir, "IMG_0002.jpg")
		createTestJPEG(t, valid, "2024:06:01 09:00:00")
		createJPEGNoExif(t, invalid)

		plan := buildPlan([]string{invalid, valid}, false, "")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		if plan[0].src != valid {
			t.Errorf("src = %q, want %q", plan[0].src, valid)
		}
	})
}

func TestExecute(t *testing.T) {
	t.Run("dry run does not rename files", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "IMG_0001.jpg")
		dst := filepath.Join(dir, "2024_03_15_1430_00.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		count := execute([]renamePair{{src: src, dst: dst}}, true)

		if count != 1 {
			t.Errorf("count = %d, want 1", count)
		}
		if !fileExists(src) {
			t.Error("source should still exist after dry run")
		}
		if fileExists(dst) {
			t.Error("destination should not exist after dry run")
		}
	})

	t.Run("actual rename moves files", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "IMG_0001.jpg")
		dst := filepath.Join(dir, "2024_03_15_1430_00.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		count := execute([]renamePair{{src: src, dst: dst}}, false)

		if count != 1 {
			t.Errorf("count = %d, want 1", count)
		}
		if fileExists(src) {
			t.Error("source should not exist after rename")
		}
		if !fileExists(dst) {
			t.Error("destination should exist after rename")
		}
	})

	t.Run("multiple files renamed", func(t *testing.T) {
		dir := t.TempDir()
		pairs := []renamePair{
			{
				src: filepath.Join(dir, "a.jpg"),
				dst: filepath.Join(dir, "2024_03_15_1430_00.jpg"),
			},
			{
				src: filepath.Join(dir, "b.jpg"),
				dst: filepath.Join(dir, "2024_03_15_1430_01.jpg"),
			},
		}
		createTestJPEG(t, pairs[0].src, "2024:03:15 14:30:00")
		createTestJPEG(t, pairs[1].src, "2024:03:15 14:30:00")

		count := execute(pairs, false)

		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
		for _, p := range pairs {
			if fileExists(p.src) {
				t.Errorf("source %q should not exist", filepath.Base(p.src))
			}
			if !fileExists(p.dst) {
				t.Errorf("destination %q should exist", filepath.Base(p.dst))
			}
		}
	})

	t.Run("empty plan", func(t *testing.T) {
		count := execute(nil, false)
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
	})
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	existing := filepath.Join(dir, "exists.txt")
	os.WriteFile(existing, []byte("hello"), 0644)

	if !fileExists(existing) {
		t.Error("fileExists() = false for existing file")
	}
	if fileExists(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("fileExists() = true for non-existing file")
	}
}

func TestParseExifMakeModel(t *testing.T) {
	t.Run("reads Make and Model from EXIF", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jpg")
		createTestJPEGFull(t, path, "2024:03:15 14:30:00", "SONY", "ZV-E10M2")

		info, err := parseExif(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.make != "SONY" {
			t.Errorf("make = %q, want %q", info.make, "SONY")
		}
		if info.model != "ZV-E10M2" {
			t.Errorf("model = %q, want %q", info.model, "ZV-E10M2")
		}
		want := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
		if !info.date.Equal(want) {
			t.Errorf("date = %v, want %v", info.date, want)
		}
	})

	t.Run("Make/Model empty when not present", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.jpg")
		createTestJPEG(t, path, "2024:03:15 14:30:00")

		info, err := parseExif(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.make != "" {
			t.Errorf("make = %q, want empty", info.make)
		}
		if info.model != "" {
			t.Errorf("model = %q, want empty", info.model)
		}
	})
}

func TestResolveSuffix(t *testing.T) {
	info := exifInfo{
		date:  time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
		make:  "SONY",
		model: "ZV-E10M2",
	}

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{"empty template", "", ""},
		{"literal suffix", "EOS", "EOS"},
		{"model template", "%model", "ZV-E10M2"},
		{"make template", "%make", "SONY"},
		{"combined", "%make_%model", "SONY_ZV-E10M2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSuffix(tt.tmpl, info)
			if got != tt.want {
				t.Errorf("resolveSuffix(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}

	t.Run("make with spaces is sanitized", func(t *testing.T) {
		infoSpaces := exifInfo{
			date:  time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
			make:  "NIKON CORPORATION",
			model: "NIKON Z 50",
		}
		got := resolveSuffix("%make_%model", infoSpaces)
		if got != "NIKON-CORPORATION_NIKON-Z-50" {
			t.Errorf("got %q, want %q", got, "NIKON-CORPORATION_NIKON-Z-50")
		}
	})

	t.Run("empty Make/Model in template", func(t *testing.T) {
		infoEmpty := exifInfo{
			date: time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
		}
		got := resolveSuffix("%model", infoEmpty)
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestSanitizeSuffix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"EOS", "EOS"},
		{"ZV-E10M2", "ZV-E10M2"},
		{"NIKON Z 50", "NIKON-Z-50"},
		{"Canon/EOS", "CanonEOS"},
		{"a<b>c:d", "abcd"},
		{"hello world!", "hello-world"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeSuffix(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeSuffix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildPlanWithSuffix(t *testing.T) {
	t.Run("manual suffix appended to filename", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "IMG_0001.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		plan := buildPlan([]string{src}, false, "EOS")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		wantDst := filepath.Join(dir, "2024_03_15_1430_00_EOS.jpg")
		if plan[0].dst != wantDst {
			t.Errorf("dst = %q, want %q", plan[0].dst, wantDst)
		}
	})

	t.Run("model template resolved from EXIF", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "IMG_0001.jpg")
		createTestJPEGFull(t, src, "2024:03:15 14:30:00", "SONY", "ZV-E10M2")

		plan := buildPlan([]string{src}, false, "%model")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		wantDst := filepath.Join(dir, "2024_03_15_1430_00_ZV-E10M2.jpg")
		if plan[0].dst != wantDst {
			t.Errorf("dst = %q, want %q", plan[0].dst, wantDst)
		}
	})

	t.Run("duplicate timestamps with suffix get sequential numbers", func(t *testing.T) {
		dir := t.TempDir()
		files := []string{
			filepath.Join(dir, "a.jpg"),
			filepath.Join(dir, "b.jpg"),
		}
		for _, f := range files {
			createTestJPEG(t, f, "2024:03:15 14:30:00")
		}

		plan := buildPlan(files, false, "EOS")

		if len(plan) != 2 {
			t.Fatalf("got %d pairs, want 2", len(plan))
		}
		if filepath.Base(plan[0].dst) != "2024_03_15_1430_00_EOS.jpg" {
			t.Errorf("plan[0] dst = %q, want 2024_03_15_1430_00_EOS.jpg", filepath.Base(plan[0].dst))
		}
		if filepath.Base(plan[1].dst) != "2024_03_15_1430_01_EOS.jpg" {
			t.Errorf("plan[1] dst = %q, want 2024_03_15_1430_01_EOS.jpg", filepath.Base(plan[1].dst))
		}
	})

	t.Run("disk collision with suffix increments sequence", func(t *testing.T) {
		dir := t.TempDir()
		collision := filepath.Join(dir, "2024_03_15_1430_00_EOS.jpg")
		os.WriteFile(collision, []byte("existing"), 0644)

		src := filepath.Join(dir, "IMG_0001.jpg")
		createTestJPEG(t, src, "2024:03:15 14:30:00")

		plan := buildPlan([]string{src}, false, "EOS")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		wantDst := filepath.Join(dir, "2024_03_15_1430_01_EOS.jpg")
		if plan[0].dst != wantDst {
			t.Errorf("dst = %q, want %q", plan[0].dst, wantDst)
		}
	})

	t.Run("make_model combined template", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "IMG_0001.jpg")
		createTestJPEGFull(t, src, "2024:03:15 14:30:00", "SONY", "ZV-E10M2")

		plan := buildPlan([]string{src}, false, "%make_%model")

		if len(plan) != 1 {
			t.Fatalf("got %d pairs, want 1", len(plan))
		}
		wantDst := filepath.Join(dir, "2024_03_15_1430_00_SONY_ZV-E10M2.jpg")
		if plan[0].dst != wantDst {
			t.Errorf("dst = %q, want %q", plan[0].dst, wantDst)
		}
	})
}
