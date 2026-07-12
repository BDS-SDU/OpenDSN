package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	previewMaxTextSize = 10 * 1024 * 1024
	previewCacheDir    = ".opendsn_preview_cache"
)

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dir := r.URL.Query().Get("dir")
	if strings.TrimSpace(dir) == "" {
		dir = "/"
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid directory path")
		return
	}

	info, err := os.Stat(absDir)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot access directory")
		return
	}
	if !info.IsDir() {
		writeError(w, http.StatusBadRequest, "path is not a directory")
		return
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot read directory")
		return
	}

	var result []BrowseEntry
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, BrowseEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  entryInfo.Size(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	writeJSON(w, http.StatusOK, BrowseResponse{
		Path:    absDir,
		Entries: result,
	})
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	absPath, info, err := resolvePreviewTarget(r.URL.Query().Get("file"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	if isWordDocument(ext) {
		pdfPath, err := s.ensurePreviewPDF(absPath, info)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, PreviewResponse{
			Filename:    filepath.Base(absPath),
			Kind:        "pdf",
			ResourceURL: "/api/preview/raw?file=" + url.QueryEscape(pdfPath),
		})
		return
	}

	if ext == ".pdf" {
		writeJSON(w, http.StatusOK, PreviewResponse{
			Filename:    filepath.Base(absPath),
			Kind:        "pdf",
			ResourceURL: "/api/preview/raw?file=" + url.QueryEscape(absPath),
		})
		return
	}

	content, err := readPreviewText(absPath, info.Size())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PreviewResponse{
		Filename: filepath.Base(absPath),
		Kind:     "text",
		Content:  content,
	})
}

func (s *Server) handlePreviewRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	absPath, _, err := resolvePreviewTarget(r.URL.Query().Get("file"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		if ext == ".pdf" {
			contentType = "application/pdf"
		} else {
			contentType = "application/octet-stream"
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(absPath)))
	http.ServeFile(w, r, absPath)
}

func resolvePreviewTarget(fileParam string) (string, os.FileInfo, error) {
	if strings.TrimSpace(fileParam) == "" {
		return "", nil, fmt.Errorf("file parameter is required")
	}

	absPath, err := filepath.Abs(fileParam)
	if err != nil {
		return "", nil, fmt.Errorf("invalid file path")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", nil, fmt.Errorf("cannot access file")
	}
	if info.IsDir() {
		return "", nil, fmt.Errorf("path is a directory, not a file")
	}

	return absPath, info, nil
}

func isWordDocument(ext string) bool {
	switch ext {
	case ".doc", ".docx":
		return true
	default:
		return false
	}
}

func readPreviewText(absPath string, size int64) (string, error) {
	if size > previewMaxTextSize {
		return "", fmt.Errorf("file too large (max 10MB) for text preview")
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot open file")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, previewMaxTextSize+1))
	if err != nil {
		return "", fmt.Errorf("cannot read file")
	}

	if isLikelyBinary(data) {
		return "", fmt.Errorf("preview is supported for text, PDF, .doc, and .docx files")
	}

	return string(data), nil
}

func isLikelyBinary(data []byte) bool {
	limit := len(data)
	if limit > 4096 {
		limit = 4096
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

func (s *Server) ensurePreviewPDF(sourcePath string, info os.FileInfo) (string, error) {
	cacheRoot := filepath.Join(s.cfg.RepoRoot, previewCacheDir)
	if err := os.MkdirAll(cacheRoot, 0755); err != nil {
		return "", fmt.Errorf("create preview cache: %w", err)
	}

	key := previewCacheKey(sourcePath, info)
	cachedPDF := filepath.Join(cacheRoot, key+".pdf")
	if _, err := os.Stat(cachedPDF); err == nil {
		return cachedPDF, nil
	}

	converter, err := lookupLibreOfficeBinary()
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp(cacheRoot, "doc-preview-")
	if err != nil {
		return "", fmt.Errorf("create preview temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command(converter, "--headless", "--convert-to", "pdf", "--outdir", tempDir, sourcePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("convert word document to pdf: %v (%s)", err, strings.TrimSpace(string(out)))
	}

	generatedPDF := filepath.Join(tempDir, strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))+".pdf")
	if _, err := os.Stat(generatedPDF); err != nil {
		return "", fmt.Errorf("converted pdf not found")
	}

	if err := moveOrCopyFile(generatedPDF, cachedPDF); err != nil {
		return "", fmt.Errorf("store preview pdf: %w", err)
	}

	return cachedPDF, nil
}

func previewCacheKey(sourcePath string, info os.FileInfo) string {
	normalized := fmt.Sprintf("%s|%d|%d", sourcePath, info.ModTime().UnixNano(), info.Size())
	digest := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(digest[:16])
}

func lookupLibreOfficeBinary() (string, error) {
	candidates := []string{"libreoffice", "soffice"}
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("libreoffice is required for Word preview but was not found in PATH")
}

func moveOrCopyFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}

	return output.Close()
}
