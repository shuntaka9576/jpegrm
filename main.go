package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

var (
	dryRun    = flag.Bool("n", false, "プレビューのみ（実際にはリネームしない）")
	recursive = flag.Bool("r", false, "サブディレクトリも走査")
	verbose   = flag.Bool("v", false, "スキップしたファイル等の詳細表示")
)

var suffix string

func init() {
	flag.StringVar(&suffix, "s", "", "サフィックス（%make, %model でEXIF値に置換）")
	flag.StringVar(&suffix, "suffix", "", "サフィックス（%make, %model でEXIF値に置換）")
}

type exifInfo struct {
	date  time.Time
	make  string
	model string
}

func parseExif(path string) (exifInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return exifInfo{}, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return exifInfo{}, err
	}

	var dt time.Time
	found := false
	for _, tag := range []exif.FieldName{exif.DateTimeOriginal, exif.DateTimeDigitized, exif.DateTime} {
		t, err := x.Get(tag)
		if err != nil {
			continue
		}
		s := strings.Trim(t.String(), "\"")
		parsed, err := time.Parse("2006:01:02 15:04:05", s)
		if err != nil {
			continue
		}
		dt = parsed
		found = true
		break
	}
	if !found {
		return exifInfo{}, fmt.Errorf("no date tag")
	}

	info := exifInfo{date: dt}
	if t, err := x.Get(exif.Make); err == nil {
		info.make = strings.TrimSpace(strings.Trim(t.String(), "\""))
	}
	if t, err := x.Get(exif.Model); err == nil {
		info.model = strings.TrimSpace(strings.Trim(t.String(), "\""))
	}

	return info, nil
}

func resolveSuffix(tmpl string, info exifInfo) string {
	if tmpl == "" {
		return ""
	}
	s := tmpl
	s = strings.ReplaceAll(s, "%make", info.make)
	s = strings.ReplaceAll(s, "%model", info.model)
	return sanitizeSuffix(s)
}

func sanitizeSuffix(s string) string {
	s = strings.ReplaceAll(s, " ", "-")
	var buf strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func isJPEG(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg"
}

func normalizePattern(pattern string) string {
	if pattern == "" || pattern == "*" || pattern == "*.*" {
		return "*"
	}
	if !strings.ContainsAny(pattern, "*?[") && !strings.Contains(pattern, ".") {
		return pattern + ".*"
	}
	return pattern
}

func matchesPattern(filename, pattern string) bool {
	if pattern == "*" {
		return true
	}
	matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(filename))
	return matched
}

func collectFiles(dir string, rec bool, pattern string) ([]string, error) {
	var files []string
	if rec {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && isJPEG(info.Name()) && matchesPattern(info.Name(), pattern) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && isJPEG(e.Name()) && matchesPattern(e.Name(), pattern) {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

type renamePair struct {
	src string
	dst string
}

func buildPlan(files []string, verb bool, sfx string) []renamePair {
	var plan []renamePair
	nameCount := map[string]int{}

	for _, src := range files {
		info, err := parseExif(src)
		if err != nil {
			if verb {
				fmt.Fprintf(os.Stderr, "SKIP: %s: %v\n", filepath.Base(src), err)
			}
			continue
		}

		base := info.date.Format("2006_01_02_1504")
		resolved := resolveSuffix(sfx, info)
		dir := filepath.Dir(src)

		key := base
		if resolved != "" {
			key = base + "_" + resolved
		}

		if _, ok := nameCount[key]; !ok {
			nameCount[key] = 0
		} else {
			nameCount[key]++
		}

		var candidate string
		if resolved != "" {
			candidate = fmt.Sprintf("%s_%02d_%s.jpg", base, nameCount[key], resolved)
		} else {
			candidate = fmt.Sprintf("%s_%02d.jpg", base, nameCount[key])
		}
		dst := filepath.Join(dir, candidate)
		for dst != src && fileExists(dst) {
			nameCount[key]++
			if resolved != "" {
				candidate = fmt.Sprintf("%s_%02d_%s.jpg", base, nameCount[key], resolved)
			} else {
				candidate = fmt.Sprintf("%s_%02d.jpg", base, nameCount[key])
			}
			dst = filepath.Join(dir, candidate)
		}

		if dst == src {
			if verb {
				fmt.Fprintf(os.Stderr, "SKIP: Already named correctly: %s\n", filepath.Base(src))
			}
			continue
		}

		plan = append(plan, renamePair{src: src, dst: dst})
	}

	return plan
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func execute(plan []renamePair, dry bool) int {
	prefix := ""
	if dry {
		prefix = "[DRY RUN] "
	}
	for _, p := range plan {
		srcDir := filepath.Dir(p.src)
		dstDir := filepath.Dir(p.dst)
		if srcDir == dstDir {
			fmt.Printf("%s%s -> %s\n", prefix, filepath.Base(p.src), filepath.Base(p.dst))
		} else {
			fmt.Printf("%s%s -> %s\n", prefix, p.src, p.dst)
		}
		if !dry {
			if err := os.Rename(p.src, p.dst); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", filepath.Base(p.src), err)
			}
		}
	}
	return len(plan)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [path]\n\nEXIF撮影日時でJPEGファイルをリネーム (YYYY_MM_DD_HHMM_NN.jpg)\n\nArguments:\n  path  対象ディレクトリ、またはディレクトリ+パターン (省略時: カレントディレクトリ)\n        例: C:\\Photos          全JPEGファイル\n            C:\\Photos\\*.*      全JPEGファイル\n            C:\\Photos\\DSC1234  DSC1234.jpg のみ\n            C:\\Photos\\DSC*     DSC で始まるファイル\n\nSuffix templates:\n  %%make   カメラメーカー (例: SONY, Canon)\n  %%model  カメラ機種名 (例: ZV-E10M2, EOS R5)\n  例: -s %%model  → 2024_03_15_1430_00_ZV-E10M2.jpg\n      -s EOS     → 2024_03_15_1430_00_EOS.jpg\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	dir := "."
	pattern := "*"
	if flag.NArg() >= 1 {
		arg := flag.Arg(0)
		info, err := os.Stat(arg)
		if err == nil && info.IsDir() {
			dir = arg
		} else {
			dir = filepath.Dir(arg)
			pattern = normalizePattern(filepath.Base(arg))
		}
	}

	if _, err := filepath.Match(pattern, "test"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid pattern: %s\n", pattern)
		os.Exit(1)
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "ERROR: Directory not found: %s\n", dir)
		os.Exit(1)
	}

	files, err := collectFiles(dir, *recursive, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		if pattern != "*" {
			fmt.Printf("No JPEG files matching pattern '%s' found.\n", pattern)
		} else {
			fmt.Println("No JPEG files found.")
		}
		return
	}

	plan := buildPlan(files, *verbose, suffix)
	if len(plan) == 0 {
		fmt.Println("No files to rename.")
		return
	}

	count := execute(plan, *dryRun)

	if *dryRun {
		fmt.Printf("\nDry run complete. %d files would be renamed.\n", count)
	} else {
		fmt.Printf("\nRenamed %d files.\n", count)
	}
}
