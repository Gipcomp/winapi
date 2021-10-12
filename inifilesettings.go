// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Gipcomp/winapi/errs"
)

const iniFileTimeStampFormat = "2006-01-02"

type IniFileSettings struct {
	fileName       string
	key2Record     map[string]iniFileRecord
	expireDuration time.Duration
	portable       bool
}

type iniFileRecord struct {
	value     string
	timestamp time.Time
}

func NewIniFileSettings(fileName string) *IniFileSettings {
	return &IniFileSettings{
		fileName:   fileName,
		key2Record: make(map[string]iniFileRecord),
	}
}

func (ifs *IniFileSettings) Get(key string) (string, bool) {
	record, ok := ifs.key2Record[key]
	return record.value, ok
}

func (ifs *IniFileSettings) Timestamp(key string) (time.Time, bool) {
	record, ok := ifs.key2Record[key]
	return record.timestamp, ok
}

func (ifs *IniFileSettings) Put(key, value string) error {
	return ifs.put(key, value, false)
}

func (ifs *IniFileSettings) PutExpiring(key, value string) error {
	return ifs.put(key, value, true)
}

func (ifs *IniFileSettings) put(key, value string, expiring bool) error {
	if key == "" {
		return errs.NewError("key must not be empty")
	}
	// if strings.IndexAny(key, "|=\r\n") > -1 {
	if strings.ContainsAny(key, "|=\r\n") {
		return errs.NewError("key contains at least one of the invalid characters '|=\\r\\n'")
	}
	// if strings.IndexAny(value, "\r\n") > -1 {
	if strings.ContainsAny(value, "\r\n") {
		return errs.NewError("value contains at least one of the invalid characters '\\r\\n'")
	}

	var timestamp time.Time
	if expiring {
		timestamp = time.Now()
	}

	ifs.key2Record[key] = iniFileRecord{value, timestamp}

	return nil
}

func (ifs *IniFileSettings) Remove(key string) error {
	delete(ifs.key2Record, key)

	return nil
}

func (ifs *IniFileSettings) ExpireDuration() time.Duration {
	return ifs.expireDuration
}

func (ifs *IniFileSettings) SetExpireDuration(expireDuration time.Duration) {
	ifs.expireDuration = expireDuration
}

func (ifs *IniFileSettings) Portable() bool {
	return ifs.portable
}

func (ifs *IniFileSettings) SetPortable(portable bool) {
	ifs.portable = portable
}

func (ifs *IniFileSettings) FilePath() string {
	if ifs.portable {
		absPath, err := filepath.Abs(ifs.fileName)
		if err != nil {
			return ""
		}

		return absPath
	}

	appDataPath, err := AppDataPath()
	if err != nil {
		return ""
	}

	return filepath.Join(
		appDataPath,
		App().OrganizationName(),
		App().ProductName(),
		ifs.fileName)
}

func (ifs *IniFileSettings) fileExists() (bool, error) {
	filePath := ifs.FilePath()

	if _, err := os.Stat(filePath); err != nil {
		// FIXME: Not necessarily a file does not exist error.
		return false, nil
	}

	return true, nil
}

func (ifs *IniFileSettings) withFile(flags int, f func(file *os.File) error) error {
	filePath := ifs.FilePath()

	dirPath, _ := filepath.Split(filePath)
	if err := os.MkdirAll(dirPath, 0644); err != nil {
		return errs.WrapError(err)
	}

	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return errs.WrapError(err)
	}
	defer file.Close()

	return f(file)
}

func (ifs *IniFileSettings) Load() error {
	exists, err := ifs.fileExists()
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return ifs.withFile(os.O_RDONLY, func(file *os.File) error {
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			assignIndex := strings.Index(line, "=")
			if assignIndex == -1 {
				return errs.NewError("bad line format: missing '='")
			}

			key := strings.TrimSpace(line[:assignIndex])

			var ts time.Time
			if parts := strings.Split(key, "|"); len(parts) > 1 {
				key = parts[0]
				if ts, _ = time.Parse(iniFileTimeStampFormat, parts[1]); ts.IsZero() {
					ts = time.Now()
				}
			}

			value := strings.TrimSpace(line[assignIndex+1:])

			ifs.key2Record[key] = iniFileRecord{value, ts}
		}

		return scanner.Err()
	})
}

func (ifs *IniFileSettings) Save() error {
	return ifs.withFile(os.O_CREATE|os.O_TRUNC|os.O_WRONLY, func(file *os.File) error {
		bufWriter := bufio.NewWriter(file)

		keys := make([]string, 0, len(ifs.key2Record))

		for key, record := range ifs.key2Record {
			if ifs.expireDuration <= 0 || record.timestamp.IsZero() || time.Since(record.timestamp) < ifs.expireDuration {
				keys = append(keys, key)
			}
		}

		sort.Strings(keys)

		for _, key := range keys {
			record := ifs.key2Record[key]

			if _, err := bufWriter.WriteString(key); err != nil {
				return errs.WrapError(err)
			}
			if !record.timestamp.IsZero() {
				if _, err := bufWriter.WriteString("|"); err != nil {
					return errs.WrapError(err)
				}
				if _, err := bufWriter.WriteString(record.timestamp.Format(iniFileTimeStampFormat)); err != nil {
					return errs.WrapError(err)
				}
			}
			if _, err := bufWriter.WriteString("="); err != nil {
				return errs.WrapError(err)
			}
			if _, err := bufWriter.WriteString(record.value); err != nil {
				return errs.WrapError(err)
			}
			if _, err := bufWriter.WriteString("\r\n"); err != nil {
				return errs.WrapError(err)
			}
		}

		return bufWriter.Flush()
	})
}
