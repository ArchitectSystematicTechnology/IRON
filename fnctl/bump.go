package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	bumper "github.com/giantswarm/semver-bump/bump"
	"github.com/giantswarm/semver-bump/storage"
	"github.com/urfave/cli"
)

var (
	initialVersion = "0.0.1"

	errVersionFileNotFound = errors.New("no VERSION file found for this function")
)

func bump() cli.Command {
	cmd := bumpcmd{commoncmd: &commoncmd{}}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:   "bump",
		Usage:  "bump function version",
		Flags:  flags,
		Action: cmd.scan,
	}
}

type bumpcmd struct {
	*commoncmd
}

func (b *bumpcmd) scan(c *cli.Context) error {
	scan(b.verbose, b.wd, b.walker)
	return nil
}

func (b *bumpcmd) walker(path string, info os.FileInfo, err error, w io.Writer) error {
	walker(path, info, err, w, b.bump)
	return nil
}

// bump will take the found valid function and bump its version
func (b *bumpcmd) bump(path string) error {
	fmt.Fprintln(verbwriter, "bumping", path)

	dir := filepath.Dir(path)
	versionfile := filepath.Join(dir, "VERSION")
	if _, err := os.Stat(versionfile); os.IsNotExist(err) {
		return errVersionFileNotFound
	}

	s, err := storage.NewVersionStorage("file", initialVersion)
	if err != nil {
		return err
	}

	version := bumper.NewSemverBumper(s, versionfile)
	newver, err := version.BumpPatchVersion("", "")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(versionfile, []byte(newver.String()), 0666); err != nil {
		return err
	}

	return nil
}
