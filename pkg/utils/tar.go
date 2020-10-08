package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileToSave struct {
	Location    string
	Destination string
}

func CreateTar(tarName string, paths []FileToSave) error {
	tarFile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	defer tarFile.Close()
	tw := tar.NewWriter(tarFile)
	defer tw.Close()
	for _, path := range paths {
		walker := func(f string, fi os.FileInfo, err error) error {
			hdr, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}
			relFilePath := f
			if filepath.IsAbs(path.Location) {
				relFilePath, err = filepath.Rel(path.Location, f)
				if err != nil {
					return err
				}
			}
			hdr.Name = filepath.Join(path.Destination, relFilePath)
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if fi.Mode().IsDir() {
				return nil
			}
			srcFile, err := os.Open(f)
			defer srcFile.Close()
			_, err = io.Copy(tw, srcFile)
			if err != nil {
				return err
			}
			return nil
		}
		if err := filepath.Walk(path.Location, walker); err != nil {
			fmt.Printf("failed to add %s to tar: %s\n", path.Location, err)
		}
	}
	return nil
}
