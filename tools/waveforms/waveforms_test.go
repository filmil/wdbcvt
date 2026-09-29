// SPDX-License-Identifier: Apache-2.0

package waveforms

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func findContribWDB(t *testing.T) string {
	t.Helper()
	// Check runfiles first, then workspace relative.
	candidates := []string{
		filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"), "testdata/contrib/gh_pr1/repro_0001_record_trailer_arp_tb.wdb"),
		"testdata/contrib/gh_pr1/repro_0001_record_trailer_arp_tb.wdb",
		"../../testdata/contrib/gh_pr1/repro_0001_record_trailer_arp_tb.wdb",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Fatalf("could not find test wdb in candidates: %v", candidates)
	return ""
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatal(err)
	}
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
}

func TestBundle(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	binDir := filepath.Join(tmp, "bin")
	outArchive := filepath.Join(tmp, "out", "waveforms.tar.gz")

	sampleWDB := findContribWDB(t)

	// 1. Setup mock bin directory with sim.wdb, sim.vcd, and xsim.log.
	simWDB := filepath.Join(binDir, "hdl", "counter", "sim.wdb")
	copyFile(t, sampleWDB, simWDB)

	simVCD := filepath.Join(binDir, "hdl", "counter", "sim.vcd")
	if err := os.WriteFile(simVCD, []byte("$timescale 1ps $end\n"), 0644); err != nil {
		t.Fatal(err)
	}
	simLog := filepath.Join(binDir, "hdl", "counter", "simulate.log")
	if err := os.WriteFile(simLog, []byte("Vivado log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Setup mock src directory with HDL sources and junk AMD/Vivado files.
	vhdlFile := filepath.Join(srcDir, "hdl", "counter", "counter.ent.vhdl")
	copyFile(t, sampleWDB, vhdlFile) // dummy content is fine for source
	if err := os.WriteFile(vhdlFile, []byte("entity counter is end;"), 0644); err != nil {
		t.Fatal(err)
	}
	vlogFile := filepath.Join(srcDir, "hdl", "counter", "tb.v")
	if err := os.WriteFile(vlogFile, []byte("module tb; endmodule"), 0644); err != nil {
		t.Fatal(err)
	}

	junkJou := filepath.Join(srcDir, "hdl", "counter", "xsim.jou")
	if err := os.WriteFile(junkJou, []byte("journal"), 0644); err != nil {
		t.Fatal(err)
	}
	junkDir := filepath.Join(srcDir, "hdl", "counter", "xsim.dir")
	if err := os.MkdirAll(junkDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(junkDir, "xelab.pb"), []byte("pb"), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Run Bundle.
	opts := Options{
		BinDir:  binDir,
		SrcDir:  srcDir,
		OutPath: outArchive,
		Prefix:  "wdbcvt-waveforms/",
	}
	if err := Bundle(opts); err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// 4. Verify archive contents.
	f, err := os.Open(outArchive)
	if err != nil {
		t.Fatalf("opening archive: %v", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("creating gzip reader: %v", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	found := make(map[string]bool)
	var prev string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading tar entry: %v", err)
		}

		if prev != "" && strings.Compare(prev, hdr.Name) >= 0 {
			t.Errorf("archive entries not strictly sorted: %q after %q", hdr.Name, prev)
		}
		prev = hdr.Name

		found[hdr.Name] = true

		if isVivadoAMD(hdr.Name) {
			t.Errorf("archive contains Vivado/AMD file: %s", hdr.Name)
		}
	}

	expected := []string{
		"wdbcvt-waveforms/hdl/counter/counter.ent.vhdl",
		"wdbcvt-waveforms/hdl/counter/sim.fst",
		"wdbcvt-waveforms/hdl/counter/sim.wdb",
		"wdbcvt-waveforms/hdl/counter/tb.v",
	}

	for _, exp := range expected {
		if !found[exp] {
			t.Errorf("missing expected entry %s in archive (found: %v)", exp, found)
		}
	}

	forbidden := []string{
		"wdbcvt-waveforms/hdl/counter/sim.vcd",
		"wdbcvt-waveforms/hdl/counter/simulate.log",
		"wdbcvt-waveforms/hdl/counter/xsim.jou",
		"wdbcvt-waveforms/hdl/counter/xsim.dir/xelab.pb",
	}

	for _, forb := range forbidden {
		if found[forb] {
			t.Errorf("found forbidden entry %s in archive", forb)
		}
	}
}
