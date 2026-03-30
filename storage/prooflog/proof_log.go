package prooflog

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const fileName = "proof_info.log"

type ProofType string
type ProofStatus string

const (
	StorageProof ProofType = "storage_proof"
	WindowPoSt   ProofType = "window_post"

	StatusProcessing ProofStatus = "processing"
	StatusCompleted  ProofStatus = "completed"
)

type ProofRecord struct {
	Type                       ProofType
	Status                     ProofStatus
	Timestamp                  string
	GenerateDurationSeconds    string
	VerifyDurationMilliseconds string
}

type proofInfoState struct {
	StorageProof ProofRecord
	WindowPoSt   ProofRecord
}

var mu sync.Mutex

func StartProof(pt ProofType) (time.Time, error) {
	now := time.Now()

	rec := ProofRecord{
		Type:                       pt,
		Status:                     StatusProcessing,
		Timestamp:                  now.Format(time.RFC3339),
		GenerateDurationSeconds:    "0",
		VerifyDurationMilliseconds: "0",
	}

	mu.Lock()
	defer mu.Unlock()

	st, err := loadState(fileName)
	if err != nil {
		return time.Time{}, err
	}

	setRecord(&st, rec)

	if err := writeState(fileName, st); err != nil {
		return time.Time{}, err
	}

	return now, nil
}

func CompleteProof(pt ProofType, startedAt time.Time, verifyStartedAt time.Time) error {
	now := time.Now()

	rec := ProofRecord{
		Type:                       pt,
		Status:                     StatusCompleted,
		Timestamp:                  now.Format(time.RFC3339),
		GenerateDurationSeconds:    formatSeconds(verifyStartedAt.Sub(startedAt)),
		VerifyDurationMilliseconds: formatMilliseconds(now.Sub(verifyStartedAt)),
	}

	mu.Lock()
	defer mu.Unlock()

	st, err := loadState(fileName)
	if err != nil {
		return err
	}

	setRecord(&st, rec)

	return writeState(fileName, st)
}

func setRecord(st *proofInfoState, rec ProofRecord) {
	switch rec.Type {
	case StorageProof:
		st.StorageProof = rec
	case WindowPoSt:
		st.WindowPoSt = rec
	}
}

func loadState(path string) (proofInfoState, error) {
	st := proofInfoState{
		StorageProof: ProofRecord{
			Type:                       StorageProof,
			GenerateDurationSeconds:    "0",
			VerifyDurationMilliseconds: "0",
		},
		WindowPoSt: ProofRecord{
			Type:                       WindowPoSt,
			GenerateDurationSeconds:    "0",
			VerifyDurationMilliseconds: "0",
		},
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "storage_proof_type":
			st.StorageProof.Type = ProofType(val)
		case "storage_proof_status":
			st.StorageProof.Status = ProofStatus(val)
		case "storage_proof_timestamp":
			st.StorageProof.Timestamp = val
		case "storage_proof_generate_duration_seconds":
			st.StorageProof.GenerateDurationSeconds = val
		case "storage_proof_verify_duration_milliseconds":
			st.StorageProof.VerifyDurationMilliseconds = val
		case "storage_proof_verify_duration_seconds":
			st.StorageProof.VerifyDurationMilliseconds = val
		case "window_post_type":
			st.WindowPoSt.Type = ProofType(val)
		case "window_post_status":
			st.WindowPoSt.Status = ProofStatus(val)
		case "window_post_timestamp":
			st.WindowPoSt.Timestamp = val
		case "window_post_generate_duration_seconds":
			st.WindowPoSt.GenerateDurationSeconds = val
		case "window_post_verify_duration_milliseconds":
			st.WindowPoSt.VerifyDurationMilliseconds = val
		case "window_post_verify_duration_seconds":
			st.WindowPoSt.VerifyDurationMilliseconds = val
		}
	}

	if err := scanner.Err(); err != nil {
		return st, fmt.Errorf("read %s: %w", path, err)
	}

	return st, nil
}

func writeState(path string, st proofInfoState) error {
	lines := []string{
		fmt.Sprintf("storage_proof_type=%s", st.StorageProof.Type),
		fmt.Sprintf("storage_proof_status=%s", st.StorageProof.Status),
		fmt.Sprintf("storage_proof_timestamp=%s", st.StorageProof.Timestamp),
		fmt.Sprintf("storage_proof_generate_duration_seconds=%s", st.StorageProof.GenerateDurationSeconds),
		fmt.Sprintf("storage_proof_verify_duration_milliseconds=%s", st.StorageProof.VerifyDurationMilliseconds),
		"",
		fmt.Sprintf("window_post_type=%s", st.WindowPoSt.Type),
		fmt.Sprintf("window_post_status=%s", st.WindowPoSt.Status),
		fmt.Sprintf("window_post_timestamp=%s", st.WindowPoSt.Timestamp),
		fmt.Sprintf("window_post_generate_duration_seconds=%s", st.WindowPoSt.GenerateDurationSeconds),
		fmt.Sprintf("window_post_verify_duration_milliseconds=%s", st.WindowPoSt.VerifyDurationMilliseconds),
	}

	data := strings.Join(lines, "\n") + "\n"

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("write temp %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}

	return nil
}

func formatSeconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 3, 64)
}

func formatMilliseconds(d time.Duration) string {
	return strconv.FormatFloat(float64(d.Microseconds())/1000.0, 'f', 3, 64)
}
