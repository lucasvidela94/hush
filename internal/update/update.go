package update

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const Package = "hush-secrets"

type SelfUpdater struct {
	Exec func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func NewSelfUpdater() *SelfUpdater {
	return &SelfUpdater{Exec: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}}
}

type Result struct {
	PreviousVersion string
	NewVersion      string
	Updated         bool
	Message         string
}

func (u *SelfUpdater) Run(ctx context.Context, currentVersion string) (*Result, error) {
	latest, err := u.latestVersion(ctx)
	if err != nil {
		return nil, err
	}
	if currentVersion != "dev" && compareSemver(normalizeVersion(currentVersion), normalizeVersion(latest)) >= 0 {
		return &Result{
			PreviousVersion: currentVersion,
			NewVersion:      latest,
			Message:         fmt.Sprintf("hush is already up to date (%s)", currentVersion),
		}, nil
	}
	out, err := u.Exec(ctx, "npm", "install", "-g", "--allow-scripts="+Package, Package+"@"+latest)
	if err != nil {
		return nil, fmt.Errorf("npm install -g %s@%s falló: %w\n%s", Package, latest, err, strings.TrimSpace(string(out)))
	}
	return &Result{
		PreviousVersion: currentVersion,
		NewVersion:      latest,
		Updated:         true,
		Message:         fmt.Sprintf("updated hush from %s to %s", currentVersion, latest),
	}, nil
}

func (u *SelfUpdater) latestVersion(ctx context.Context) (string, error) {
	out, err := u.Exec(ctx, "npm", "view", Package, "version")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("hush update necesita npm: instalalo o corré a mano npm i -g %s@latest", Package)
		}
		return "", fmt.Errorf("npm view %s falló: %w\n%s", Package, err, strings.TrimSpace(string(out)))
	}
	latest := strings.TrimSpace(string(out))
	if normalizeVersion(latest) == "" {
		return "", fmt.Errorf("npm devolvió una versión inválida: %q", latest)
	}
	return latest, nil
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
