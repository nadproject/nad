/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of NAD.
 *
 * NAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * NAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with NAD.  If not, see <https://www.gnu.org/licenses/>.
 */

// Package testutils provides utilities used in tests
package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nadproject/nad/pkg/cli/consts"
	"github.com/nadproject/nad/pkg/cli/context"
	"github.com/nadproject/nad/pkg/cli/database"
	"github.com/nadproject/nad/pkg/cli/utils"
	"github.com/pkg/errors"
)

// Login simulates a logged in user by inserting credentials in the local database
func Login(t *testing.T, ctx *context.NadCtx) {
	db := ctx.DB

	database.MustExec(t, "inserting sessionKey", db, "INSERT INTO system (key, value) VALUES (?, ?)", consts.SystemSessionKey, "someSessionKey")
	database.MustExec(t, "inserting sessionKeyExpiry", db, "INSERT INTO system (key, value) VALUES (?, ?)", consts.SystemSessionKeyExpiry, time.Now().Add(24*time.Hour).Unix())

	ctx.SessionKey = "someSessionKey"
	ctx.SessionKeyExpiry = time.Now().Add(24 * time.Hour).Unix()
}

// RemoveDir cleans up the test env represented by the given context
func RemoveDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(errors.Wrap(err, "removing the directory"))
	}
}

// CopyFixture writes the content of the given fixture to the filename inside the nad dir
func CopyFixture(t *testing.T, ctx context.NadCtx, fixturePath string, filename string) {
	fp, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatal(errors.Wrap(err, "getting the absolute path for fixture"))
	}

	dp, err := filepath.Abs(filepath.Join(ctx.NADDir, filename))
	if err != nil {
		t.Fatal(errors.Wrap(err, "getting the absolute path nad dir"))
	}

	err = utils.CopyFile(fp, dp)
	if err != nil {
		t.Fatal(errors.Wrap(err, "copying the file"))
	}
}

// WriteFile writes a file with the given content and  filename inside the nad dir
func WriteFile(ctx context.NadCtx, content []byte, filename string) {
	dp, err := filepath.Abs(filepath.Join(ctx.NADDir, filename))
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(dp, content, 0644); err != nil {
		panic(err)
	}
}

// ReadFile reads the content of the file with the given name in nad dir
func ReadFile(ctx context.NadCtx, filename string) []byte {
	path := filepath.Join(ctx.NADDir, filename)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return b
}

// ReadJSON reads JSON fixture to the struct at the destination address
func ReadJSON(path string, destination interface{}) {
	var dat []byte
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(errors.Wrap(err, "Failed to load fixture payload"))
	}
	if err := json.Unmarshal(dat, destination); err != nil {
		panic(errors.Wrap(err, "Failed to get event"))
	}
}

// NewNADCmd returns a new NAD command and a pointer to stderr
func NewNADCmd(opts RunNADCmdOptions, binaryName string, arg ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer, error) {
	var stderr, stdout bytes.Buffer

	binaryPath, err := filepath.Abs(binaryName)
	if err != nil {
		return &exec.Cmd{}, &stderr, &stdout, errors.Wrap(err, "getting the absolute path to the test binary")
	}

	cmd := exec.Command(binaryPath, arg...)
	cmd.Env = []string{fmt.Sprintf("DNOTE_DIR=%s", opts.NADDir), fmt.Sprintf("DNOTE_HOME_DIR=%s", opts.HomeDir)}
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	return cmd, &stderr, &stdout, nil
}

// RunNADCmdOptions is an option for RunNADCmd
type RunNADCmdOptions struct {
	NADDir string
	HomeDir  string
}

// RunNADCmd runs a nad command
func RunNADCmd(t *testing.T, opts RunNADCmdOptions, binaryName string, arg ...string) {
	t.Logf("running: %s %s", binaryName, strings.Join(arg, " "))

	cmd, stderr, stdout, err := NewNADCmd(opts, binaryName, arg...)
	if err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrap(err, "getting command").Error())
	}

	cmd.Env = append(cmd.Env, "DNOTE_DEBUG=1")

	if err := cmd.Run(); err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrapf(err, "running command %s", stderr.String()))
	}

	// Print stdout if and only if test fails later
	t.Logf("\n%s", stdout)
}

// WaitNADCmd runs a nad command and waits until the command is exited
func WaitNADCmd(t *testing.T, opts RunNADCmdOptions, runFunc func(io.WriteCloser) error, binaryName string, arg ...string) {
	t.Logf("running: %s %s", binaryName, strings.Join(arg, " "))

	cmd, stderr, stdout, err := NewNADCmd(opts, binaryName, arg...)
	if err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrap(err, "getting command").Error())
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrap(err, "getting stdin %s"))
	}
	defer stdin.Close()

	// Start the program
	err = cmd.Start()
	if err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrap(err, "starting command"))
	}

	err = runFunc(stdin)
	if err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrap(err, "running with stdin"))
	}

	err = cmd.Wait()
	if err != nil {
		t.Logf("\n%s", stdout)
		t.Fatal(errors.Wrapf(err, "running command %s", stderr.String()))
	}

	// Print stdout if and only if test fails later
	t.Logf("\n%s", stdout)
}

// UserConfirm simulates confirmation from the user by writing to stdin
func UserConfirm(stdin io.WriteCloser) error {
	// confirm
	if _, err := io.WriteString(stdin, "y\n"); err != nil {
		return errors.Wrap(err, "indicating confirmation in stdin")
	}

	return nil
}

// MustMarshalJSON marshalls the given interface into JSON.
// If there is any error, it fails the test.
func MustMarshalJSON(t *testing.T, v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("%s: marshalling data", t.Name())
	}

	return b
}

// MustUnmarshalJSON marshalls the given interface into JSON.
// If there is any error, it fails the test.
func MustUnmarshalJSON(t *testing.T, data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		t.Fatalf("%s: unmarshalling data", t.Name())
	}
}
