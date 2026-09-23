package update

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Repo = "lucasvidela94/hush"

type SelfUpdater struct {
	HTTPClient *http.Client
	Repo       string
}

func NewSelfUpdater() *SelfUpdater {
	return &SelfUpdater{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Repo:       Repo,
	}
}

type Result struct {
	PreviousVersion string
	NewVersion      string
	Updated         bool
	Message         string
}

func (u *SelfUpdater) Run(ctx context.Context, currentVersion string) (*Result, error) {
	currentPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot locate current executable: %w", err)
	}
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve executable path: %w", err)
	}

	latest, err := u.latestVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot determine latest version: %w", err)
	}

	if currentVersion != "dev" && compareSemver(normalizeVersion(currentVersion), normalizeVersion(latest)) >= 0 {
		return &Result{
			PreviousVersion: currentVersion,
			NewVersion:      latest,
			Updated:         false,
			Message:         fmt.Sprintf("hush is already up to date (%s)", currentVersion),
		}, nil
	}

	assetName := u.assetName(latest)
	raw, err := u.downloadAndVerify(ctx, latest, assetName)
	if err != nil {
		return nil, err
	}
	if err := apply(currentPath, raw); err != nil {
		return nil, fmt.Errorf("cannot replace binary: %w", err)
	}
	return &Result{
		PreviousVersion: currentVersion,
		NewVersion:      latest,
		Updated:         true,
		Message:         fmt.Sprintf("updated hush from %s to %s", currentVersion, latest),
	}, nil
}

func compareSemver(a, b string) int {
	a = strings.TrimPrefix(strings.Split(a, "+")[0], "v")
	b = strings.TrimPrefix(strings.Split(b, "+")[0], "v")
	pa := strings.SplitN(a, "-", 2)
	pb := strings.SplitN(b, "-", 2)
	va := strings.Split(pa[0], ".")
	vb := strings.Split(pb[0], ".")
	for i := 0; i < 3 && i < len(va) && i < len(vb); i++ {
		na, _ := strconv.Atoi(va[i])
		nb, _ := strconv.Atoi(vb[i])
		if na != nb {
			if na < nb {
				return -1
			}
			return 1
		}
	}
	hasPreA := len(pa) > 1
	hasPreB := len(pb) > 1
	switch {
	case !hasPreA && hasPreB:
		return 1
	case hasPreA && !hasPreB:
		return -1
	case hasPreA && hasPreB:
		return strings.Compare(pa[1], pb[1])
	default:
		return 0
	}
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "dev" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	core := strings.Split(strings.Split(v, "+")[0], "-")[0]
	parts := strings.Split(strings.TrimPrefix(core, "v"), ".")
	if len(parts) != 3 {
		return ""
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return ""
		}
	}
	return v
}

func (u *SelfUpdater) latestVersion(ctx context.Context) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", u.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode release: %w", err)
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name in release")
	}
	return release.TagName, nil
}

func (u *SelfUpdater) assetName(version string) string {
	_ = version
	return fmt.Sprintf("hush-%s-%s.gz", runtime.GOOS, runtime.GOARCH)
}

func (u *SelfUpdater) downloadAndVerify(ctx context.Context, version, assetName string) ([]byte, error) {
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s", u.Repo, version)
	archive, err := u.download(ctx, base+"/"+assetName)
	if err != nil {
		return nil, fmt.Errorf("download asset: %w", err)
	}
	sums, err := u.download(ctx, base+"/checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("fetch checksum: %w", err)
	}
	expected, err := checksumFor(string(sums), assetName)
	if err != nil {
		return nil, err
	}
	actual := sha256.Sum256(archive)
	if !strings.EqualFold(hex.EncodeToString(actual[:]), expected) {
		return nil, fmt.Errorf("checksum mismatch for %s", assetName)
	}
	raw, err := gunzip(archive)
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	return raw, nil
}

func checksumFor(sums, assetName string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(sums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == assetName {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("checksum for %s not found", assetName)
}

func (u *SelfUpdater) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func apply(currentPath string, raw []byte) error {
	dir := filepath.Dir(currentPath)
	base := filepath.Base(currentPath)
	tmp := filepath.Join(dir, base+".new")
	if err := os.WriteFile(tmp, raw, 0o755); err != nil {
		return fmt.Errorf("write new binary: %w", err)
	}
	backup := currentPath + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(currentPath, backup); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename current binary: %w", err)
	}
	if err := os.Rename(tmp, currentPath); err != nil {
		_ = os.Rename(backup, currentPath)
		_ = os.Remove(tmp)
		return fmt.Errorf("rename new binary: %w", err)
	}
	_ = os.Remove(backup)
	return nil
}

func gunzip(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gr.Close() }()
	return io.ReadAll(gr)
}
