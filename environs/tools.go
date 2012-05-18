package environs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"launchpad.net/juju/go/log"
	"launchpad.net/juju/go/version"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var currentSeries = "precise"                // TODO find out actual version.
var currentArch = ubuntuArch(runtime.GOARCH) // TODO better

func ubuntuArch(arch string) string {
	if arch == "386" {
		arch = "i386"
	}
	return arch
}

var toolPrefix = "tools/juju-"

var toolFilePat = regexp.MustCompile(`^`+toolPrefix+`(\d+\.\d+\.\d+)-([^-]+)-([^-]+)\.tgz$`)

// toolsPathForVersion returns a path for the juju tools with the
// given version, OS and architecture.
func toolsPathForVersion(v version.Version, series, arch string) string {
	return fmt.Sprintf(toolPrefix+"%v-%s-%s.tgz", v, series, arch)
}

// ToolsPath gives the path for the current juju tools, as expected
// by environs.Environ.PutFile, for example.
var toolsPath = toolsPathForVersion(version.Current, currentSeries, currentArch)

// PutTools uploads the current version of the juju tools
// executables to the given storage.
// TODO find binaries from $PATH when go dev environment not available.
func PutTools(storage StorageWriter) error {
	// We create the entire archive before asking the environment to
	// start uploading so that we can be sure we have archived
	// correctly.
	f, err := ioutil.TempFile("", "juju-tgz")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())
	err = bundleTools(f)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	return storage.Put(toolsPath, f, fi.Size())
}

// archive writes the executable files found in the given
// directory in gzipped tar format to w.
// An error is returned if an entry inside dir is not
// a regular executable file.
func archive(w io.Writer, dir string) (err error) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	gzw := gzip.NewWriter(w)
	defer closeErrorCheck(&err, gzw)

	tarw := tar.NewWriter(gzw)
	defer closeErrorCheck(&err, tarw)

	for _, ent := range entries {
		if !isExecutable(ent) {
			return fmt.Errorf("archive: found non-executable file %q", filepath.Join(dir, ent.Name()))
		}
		h := tarHeader(ent)
		// ignore local umask
		h.Mode = 0755
		err := tarw.WriteHeader(h)
		if err != nil {
			return err
		}
		if err := readFile(tarw, filepath.Join(dir, ent.Name())); err != nil {
			return err
		}
	}
	return nil
}

// readFile writes the contents of the given file to w.
func readFile(w io.Writer, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

// tarHeader returns a tar file header given the file's stat
// information.
func tarHeader(i os.FileInfo) *tar.Header {
	return &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       i.Name(),
		Size:       i.Size(),
		Mode:       int64(i.Mode() & 0777),
		ModTime:    i.ModTime(),
		AccessTime: i.ModTime(),
		ChangeTime: i.ModTime(),
		Uname:      "ubuntu",
		Gname:      "ubuntu",
	}
}

// isExecutable returns whether the given info
// represents a regular file executable by (at least) the user.
func isExecutable(i os.FileInfo) bool {
	return i.Mode()&(0100|os.ModeType) == 0100
}

// closeErrorCheck means that we can ensure that
// Close errors do not get lost even when we defer them,
func closeErrorCheck(errp *error, c io.Closer) {
	err := c.Close()
	if *errp == nil {
		*errp = err
	}
}

// FindTools tries to find a set of tools appropriate for the current
// version and platform and returns a URL that can be used to access
// them in gzipped tar archive format.
func FindTools(env Environ) (url string, err error) {
//	storage, path, err := findTools(env)
//	if err != nil {
//		return "", err
//	}
//	return storage.URL(path), nil
	return "", fmt.Errorf("url unimplemented")
}

// GetTools finds the latest compatible version of the juju tools
// and downloads them into the given directory.
func GetTools(env Environ, dir string) error {
	storage, path, err := findTools(env)
	if err != nil {
		return err
	}
	r, err := storage.Get(path)
	if err != nil {
		return err
	}
	defer r.Close()

	r, err = gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer r.Close()

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if strings.Contains(hdr.Name, "/\\") {
			return fmt.Errorf("bad name %q in tools archive", hdr.Name)
		}

		name := filepath.Join(dir, hdr.Name)
		if err := writeFile(name, os.FileMode(hdr.Mode&0777), tr); err != nil {
			return fmt.Errorf("tar extract %q failed: %v", name, err)
		}
	}
	panic("not reached")
}

// findToolsPath is an internal version of FindTools that returns the
// found StorageReader and the path within that storage.
func findTools(env Environ) (storage StorageReader, path string, err error) {
	storage = env.Storage()
	path, err = findToolsPath(storage)
	if _, ok := err.(*NotFoundError); ok {
		storage = env.PublicStorage()
		path, err = findToolsPath(storage)
	}
	if err != nil {
		return nil, "", err
	}
	return
}

// findToolsPath looks for the tools in the given storage.
func findToolsPath(store StorageReader) (path string, err error) {
	names, err := store.List(fmt.Sprintf("%s%d.", toolPrefix, version.Current.Major))
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", &NotFoundError{fmt.Errorf("no tools found")}
	}
	bestVersion := version.Version{Major: -1}
	bestName := ""
	for _, name := range names {
		m := toolFilePat.FindStringSubmatch(name)
		if m == nil {
			log.Printf("unexpected tools file found %q", name)
			continue
		}
		vers, err := version.Parse(m[1])
		if err != nil {
			log.Printf("failed to parse version %q: %v", name, err)
			continue
		}
		if m[2] != currentSeries {
			continue
		}
		// TODO allow different architectures.
		if m[3] != currentArch {
			continue
		}
		if vers.Major != version.Current.Major {
			continue
		}
		if bestVersion.Less(vers) {
			bestVersion = vers
			bestName = name
		}
	}
	if bestVersion.Major < 0 {
		return "", &NotFoundError{fmt.Errorf("no compatible tools found")}
	}
	return bestName, nil
}

// bundleTools bundles all the current juju tools in gzipped tar
// format to the given writer.
func bundleTools(w io.Writer) error {
	dir, err := ioutil.TempDir("", "juju-tools")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	cmd := exec.Command("go", "install", "launchpad.net/juju/go/cmd/...")
	cmd.Env = []string{
		"GOPATH=" + os.Getenv("GOPATH"),
		"GOBIN=" + dir,
		"PATH=" + os.Getenv("PATH"),
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v; %s", err, out)
	}
	return archive(w, dir)
}

func writeFile(name string, mode os.FileMode, r io.Reader) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// EmptyStorage holds a StorageReader object
// that contains nothing.
var EmptyStorage StorageReader = emptyStorage{}

type emptyStorage struct{}

func (s emptyStorage) Get(name string) (io.ReadCloser, error) {
	return nil, &NotFoundError{fmt.Errorf("file %q not found in empty storage", name)}
}

func (s emptyStorage) List(prefix string) ([]string, error) {
	return nil, nil
}
