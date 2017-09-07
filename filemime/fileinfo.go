package filemime

import (
	"fmt"
	"context"
	"github.com/MXi4oyu/gomagic/magic"
	"time"
	"encoding/json"
	"strings"
	"os/exec"
)


var(
	fi FileInfo
)

// FileMagic is file magic
type FileMagic struct {
	Mime        string `json:"mime" structs:"mime"`
	Description string `json:"description" structs:"description"`
}

// FileInfo json object
type FileInfo struct {
	Magic    FileMagic         `json:"magic" structs:"magic"`
	SSDeep   string            `json:"ssdeep" structs:"ssdeep"`
}

// GetFileMimeType returns the mime-type of a file path
func GetFileMimeType(ctx context.Context, path string) error {

	c := make(chan struct {
		mimetype string
		err      error
	}, 1)

	go func() {
		magic.Open(magic.MAGIC_MIME_TYPE | magic.MAGIC_SYMLINK | magic.MAGIC_ERROR)
		defer magic.Close()

		mt, err := magic.TypeByFile(path)
		pack := struct {
			mimetype string
			err      error
		}{mt, err}
		c <- pack
	}()

	select {
	case <-ctx.Done():
		<-c // Wait for mime
		fmt.Println("Cancel the context")
		return ctx.Err()
	case ok := <-c:
		if ok.err != nil {
			fi.Magic.Mime = ok.err.Error()
			return ok.err
		}
		fi.Magic.Mime = ok.mimetype
		return nil
	}
}

// GetFileDescription returns the textual libmagic type of a file path
func GetFileDescription(ctx context.Context, path string) error {

	c := make(chan struct {
		magicdesc string
		err       error
	}, 1)

	go func() {
		magic.Open(magic.MAGIC_SYMLINK | magic.MAGIC_ERROR)
		defer magic.Close()

		magicdesc, err := magic.TypeByFile(path)
		pack := struct {
			magicdesc string
			err       error
		}{magicdesc, err}
		c <- pack
	}()

	select {
	case <-ctx.Done():
		<-c // Wait for mime
		fmt.Println("Cancel the context")
		return ctx.Err()
	case ok := <-c:
		if ok.err != nil {
			fi.Magic.Description = ok.err.Error()
			return ok.err
		}
		fi.Magic.Description = ok.magicdesc
		return nil
	}
}

func SliceContainsString(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(b, a) {
			return true
		}
	}
	return false
}

// ParseSsdeepOutput convert ssdeep output into JSON
func ParseSsdeepOutput(ssdout string, err error) string {

	if err != nil {
		return err.Error()
	}

	// Break output into lines
	lines := strings.Split(ssdout, "\n")

	if  SliceContainsString("No such file or directory", lines) {
		return ""
	}

	// Break second line into hash and path
	hashAndPath := strings.Split(lines[1], ",")

	return strings.TrimSpace(hashAndPath[0])
}


// RunCommand runs cmd on file
func RunCommand(ctx context.Context, cmd string, args ...string) (string, error) {

	var c *exec.Cmd

	if ctx != nil {
		c = exec.CommandContext(ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}

	output, err := c.Output()
	if err != nil {
		return string(output), err
	}

	// check for exec context timeout
	if ctx != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out", cmd)
		}
	}

	return string(output), nil
}


func FileInfoScan(path string)([] byte){

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	GetFileMimeType(ctx, path)
	GetFileDescription(ctx, path)

	fileInfo := FileInfo{
		Magic:fi.Magic,
		SSDeep:ParseSsdeepOutput(RunCommand(ctx, "ssdeep", path)),
	}

	fijson,_:=json.Marshal(fileInfo)

	fmt.Println(string(fijson))

	return fijson

}

