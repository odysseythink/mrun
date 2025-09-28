package crush

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ZipCompress(filenames map[string]string, output_filename string) error {
	if !strings.HasSuffix(strings.ToLower(output_filename), ".zip") {
		output_filename += ".zip"
	}
	output_zip_file, err := os.Create(output_filename)
	if err != nil {
		return err
	}
	defer output_zip_file.Close()
	output_zip_writer := zip.NewWriter(output_zip_file)
	defer output_zip_writer.Close()

	for rootOnDisk, rootInArchive := range filenames {
		walkErr := filepath.WalkDir(rootOnDisk, func(filename string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			nameInArchive := nameOnDiskToNameInArchive(filename, rootOnDisk, rootInArchive)
			// this is the root folder and we are adding its contents to target rootInArchive
			if info.IsDir() && nameInArchive == "" {
				return nil
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = nameInArchive
			header.Method = zip.Deflate
			fw, err := output_zip_writer.CreateHeader(header)
			if err != nil {
				return err
			}
			// handle symbolic links
			var linkTarget string
			if isSymlink(info) {
				// preserve symlinks
				linkTarget, err = os.Readlink(filename)
				if err != nil {
					return fmt.Errorf("%s: readlink: %w", filename, err)
				}
				_, err := fw.Write([]byte(linkTarget))
				if err != nil {
					return fmt.Errorf("%s: zip write: %w", filename, err)
				}
			} else {
				fs, err := os.Open(filename)
				if err != nil {
					return err
				}
				defer fs.Close()
				_, err = io.Copy(fw, fs)
				if err != nil {
					return fmt.Errorf("%s: copy failed: %w", filename, err)
				}
			}
			return nil
		})
		if walkErr != nil {
			return walkErr
		}
	}
	return nil
}
