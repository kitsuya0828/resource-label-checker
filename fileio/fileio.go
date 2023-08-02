package fileio

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"os"

	"gopkg.in/yaml.v3"
)

// GetStringCollection gets fields of the specified key from the data of the YAML file
func GetStringCollection(data []byte, key string) ([]string, error) {
	m := make(map[string][]string)
	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}
	collection, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("no such key: %s", key)
	}
	return collection, nil
}

// CreateZip compresses files with a specified extension in a specified directory to create a zip file
func CreateZip(inputDir string, outputFile string, ext string) error {
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext == "" || filepath.Ext(path) == ext {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			zipFile, err := zipWriter.Create(filepath.Base(path))
			if err != nil {
				return err
			}

			_, err = io.Copy(zipFile, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// DeleteAllFiles deletes all files in a specified directory
func DeleteAllFiles(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(dir, file.Name())
		err := os.Remove(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteResultToCsvFiles writes resource names to CSV files in parallel for each resource type
func WriteResultToCsvFiles(resources map[string][]string, resultDir string) error {
	if err := os.MkdirAll(resultDir, 0777); err != nil {
		return err
	}
	if err := DeleteAllFiles(resultDir); err != nil {
		return err
	}

	for k, v := range resources {
		fileName := strings.Replace(strings.Replace(k, "/", "-", -1), "::", "-", -1)
		filePath := fmt.Sprintf("%s/%s.csv", resultDir, fileName)

		var wg sync.WaitGroup
		resouceType := k
		resourceList := v
		wg.Add(1)
		go func(filePath string, resourceType string, resourceList []string, wg *sync.WaitGroup) error {
			defer wg.Done()

			file, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			records := [][]string{{"Type", "Name"}}
			for _, name := range resourceList {
				records = append(records, []string{resourceType, name})
			}
			writer := csv.NewWriter(file)
			err = writer.WriteAll(records)
			if err != nil {
				return err
			}
			return nil
		}(filePath, resouceType, resourceList, &wg)
		wg.Wait()
	}
	return nil
}
