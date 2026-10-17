package update

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v0.1.0", "v0.2.0", -1},
		{"v0.2.0", "v0.1.0", 1},
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v1.0.0-rc1", 1},
		{"0.1.22", "v0.1.22", 0},
	}
	for _, tc := range cases {
		if got := compareSemver(tc.a, tc.b); got != tc.want {
			t.Fatalf("compare(%q,%q)=%d quiero %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	if got := normalizeVersion("0.1.2"); got != "v0.1.2" {
		t.Fatalf("got %q", got)
	}
	for _, bad := range []string{"", "dev", "v1.2", "hola"} {
		if got := normalizeVersion(bad); got != "" {
			t.Fatalf("%q → %q, quiero vacío", bad, got)
		}
	}
}

type fakeNPM struct {
	latest     string
	viewErr    error
	installErr error
	calls      [][]string
}

func (f *fakeNPM) exec(_ context.Context, name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string{name}, args...))
	if len(args) > 0 && args[0] == "view" {
		return []byte(f.latest + "\n"), f.viewErr
	}
	return []byte("salida de npm"), f.installErr
}

func (f *fakeNPM) installed() bool {
	for _, c := range f.calls {
		if len(c) > 1 && c[1] == "install" {
			return true
		}
	}
	return false
}

func TestUpdateInstalaLaUltimaPorNPM(t *testing.T) {
	f := &fakeNPM{latest: "0.7.0"}
	res, err := (&SelfUpdater{Exec: f.exec}).Run(context.Background(), "v0.6.0")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated || res.NewVersion != "0.7.0" {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	last := f.calls[len(f.calls)-1]
	want := "npm install -g --allow-scripts=hush-secrets hush-secrets@0.7.0"
	if strings.Join(last, " ") != want {
		t.Fatalf("comando %q, quiero %q", strings.Join(last, " "), want)
	}
}

func TestUpdateNoHaceNadaSiYaEstaAlDia(t *testing.T) {
	for _, current := range []string{"v0.7.0", "v0.8.0"} {
		f := &fakeNPM{latest: "0.7.0"}
		res, err := (&SelfUpdater{Exec: f.exec}).Run(context.Background(), current)
		if err != nil {
			t.Fatal(err)
		}
		if res.Updated || f.installed() {
			t.Fatalf("%s: instaló sin necesidad", current)
		}
	}
}

func TestUpdateDevSiempreInstala(t *testing.T) {
	f := &fakeNPM{latest: "0.7.0"}
	res, err := (&SelfUpdater{Exec: f.exec}).Run(context.Background(), "dev")
	if err != nil || !res.Updated {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestUpdateSinNPMExplicaQueHacer(t *testing.T) {
	f := &fakeNPM{viewErr: exec.ErrNotFound}
	_, err := (&SelfUpdater{Exec: f.exec}).Run(context.Background(), "v0.6.0")
	if err == nil || !strings.Contains(err.Error(), "necesita npm") {
		t.Fatalf("error sin guía: %v", err)
	}
}

func TestUpdateFalloDeInstalacionMuestraLaSalida(t *testing.T) {
	f := &fakeNPM{latest: "0.7.0", installErr: errors.New("exit status 1")}
	_, err := (&SelfUpdater{Exec: f.exec}).Run(context.Background(), "v0.6.0")
	if err == nil || !strings.Contains(err.Error(), "salida de npm") {
		t.Fatalf("no muestra la salida de npm: %v", err)
	}
}

func TestUpdateRechazaVersionInvalidaDeNPM(t *testing.T) {
	f := &fakeNPM{latest: "no-es-version"}
	if _, err := (&SelfUpdater{Exec: f.exec}).Run(context.Background(), "v0.6.0"); err == nil {
		t.Fatal("aceptó una versión inválida")
	}
	if f.installed() {
		t.Fatal("instaló con una versión inválida")
	}
}
