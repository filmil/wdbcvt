// SPDX-License-Identifier: Apache-2.0

// Package waveforms bundles waveform databases, converted FST files,
// and HDL source files into a release archive.
package waveforms

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git.hdlfactory.com/HDL/wdbcvt/pkg/fstout"
	"git.hdlfactory.com/HDL/wdbcvt/pkg/wdb"
)

// Options configures the archive bundling.
type Options struct {
	// BinDir is the directory holding the built simulation outputs.
	BinDir string

	// SrcDir is the root directory of the source repository.
	SrcDir string

	// OutPath is the path to write the tar.gz archive to.
	OutPath string

	// Prefix is the directory prefix inside the archive.
	Prefix string
}

// archiveEntry records an entry to place in the tar archive.
type archiveEntry struct {
	archPath string
	srcPath  string
	content  []byte
}

// isHDL returns true if the file extension is a Verilog or VHDL extension.
func isHDL(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".v" || ext == ".sv" || ext == ".vhd" || ext == ".vhdl"
}

// isVivadoAMD returns true if the path belongs to an AMD or Vivado tool output.
func isVivadoAMD(path string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))
	if base == "xsim.dir" || base == ".xsim.dir" || base == "xsim.jou" || base == "xsim.log" {
		return true
	}
	if strings.HasSuffix(path, ".xsim.dir") {
		return true
	}
	if ext == ".vcd" || ext == ".jou" || ext == ".log" || ext == ".pb" || ext == ".str" {
		return true
	}
	return false
}

// Bundle collects all WDB files, converts them to FST, collects HDL sources,
// and writes the bundle archive to opts.OutPath.
func Bundle(opts Options) error {
	if opts.OutPath == "" {
		return fmt.Errorf("out path is required")
	}
	prefix := opts.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	entries := make(map[string]archiveEntry)

	// 1. Collect HDL sources from SrcDir/hdl.
	hdlRoot := filepath.Join(opts.SrcDir, "hdl")
	if info, err := os.Stat(hdlRoot); err == nil && info.IsDir() {
		err := filepath.Walk(hdlRoot, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if isVivadoAMD(p) {
					return filepath.SkipDir
				}
				return nil
			}
			if isVivadoAMD(p) {
				return nil
			}
			if isHDL(p) {
				rel, err := filepath.Rel(opts.SrcDir, p)
				if err != nil {
					return err
				}
				target := prefix + rel
				entries[target] = archiveEntry{
					archPath: target,
					srcPath:  p,
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walking hdl sources: %w", err)
		}
	}

	// 2. Collect WDB files from BinDir and SrcDir/testdata/contrib.
	wdbPaths := make(map[string]string) // relPath -> absPath

	collectWDBs := func(root, baseRel string) error {
		if root == "" {
			return nil
		}
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			return nil
		}
		return filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if isVivadoAMD(p) {
					return filepath.SkipDir
				}
				return nil
			}
			if isVivadoAMD(p) {
				return nil
			}
			if strings.EqualFold(filepath.Ext(p), ".wdb") {
				rel, err := filepath.Rel(root, p)
				if err != nil {
					return err
				}
				if baseRel != "" {
					rel = filepath.Join(baseRel, rel)
				}
				wdbPaths[rel] = p
			}
			return nil
		})
	}

	if err := collectWDBs(filepath.Join(opts.BinDir, "hdl"), "hdl"); err != nil {
		return fmt.Errorf("collecting bin wdb files: %w", err)
	}
	if err := collectWDBs(filepath.Join(opts.SrcDir, "testdata", "contrib"), filepath.Join("testdata", "contrib")); err != nil {
		return fmt.Errorf("collecting contrib wdb files: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "wdbcvt-waveforms-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for rel, abs := range wdbPaths {
		targetWDB := prefix + rel
		entries[targetWDB] = archiveEntry{
			archPath: targetWDB,
			srcPath:  abs,
		}

		// Convert to FST.
		f, err := wdb.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("reading wdb %s: %w", abs, err)
		}

		fstRel := strings.TrimSuffix(rel, filepath.Ext(rel)) + ".fst"
		targetFST := prefix + fstRel
		tmpFST := filepath.Join(tmpDir, strings.ReplaceAll(fstRel, "/", "_"))

		if err := fstout.Write(f, tmpFST); err != nil {
			return fmt.Errorf("converting %s to fst: %w", abs, err)
		}

		fstData, err := os.ReadFile(tmpFST)
		if err != nil {
			return fmt.Errorf("reading converted fst %s: %w", tmpFST, err)
		}
		os.Remove(tmpFST)

		entries[targetFST] = archiveEntry{
			archPath: targetFST,
			content:  fstData,
		}
	}

	// 3. Write deterministic tar.gz.
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := os.MkdirAll(filepath.Dir(opts.OutPath), 0755); err != nil {
		return fmt.Errorf("creating parent dir of out: %w", err)
	}

	outFile, err := os.Create(opts.OutPath)
	if err != nil {
		return fmt.Errorf("creating out archive: %w", err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	gw.Header.ModTime = time.Unix(0, 0)
	tw := tar.NewWriter(gw)

	for _, k := range keys {
		entry := entries[k]
		var size int64
		var r io.Reader

		if entry.content != nil {
			size = int64(len(entry.content))
			r = strings.NewReader(string(entry.content))
		} else {
			fi, err := os.Stat(entry.srcPath)
			if err != nil {
				return fmt.Errorf("stat %s: %w", entry.srcPath, err)
			}
			size = fi.Size()
			f, err := os.Open(entry.srcPath)
			if err != nil {
				return fmt.Errorf("open %s: %w", entry.srcPath, err)
			}
			defer f.Close()
			r = f
		}

		hdr := &tar.Header{
			Name:     entry.archPath,
			Mode:     0644,
			Size:     size,
			ModTime:  time.Unix(0, 0),
			Uid:      0,
			Gid:      0,
			Uname:    "",
			Gname:    "",
			Format:   tar.FormatPAX,
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("writing header for %s: %w", entry.archPath, err)
		}
		if _, err := io.Copy(tw, r); err != nil {
			return fmt.Errorf("writing content for %s: %w", entry.archPath, err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("closing tar writer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("closing gzip writer: %w", err)
	}
	return nil
}
