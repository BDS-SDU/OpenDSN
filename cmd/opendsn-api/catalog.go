package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	catalogDirName  = ".opendsn_catalog"
	catalogFileName = "catalog.json"
)

type Catalog struct {
	Files []CatalogFile `json:"files"`
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	catalog, err := s.readCatalog()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CatalogResponse{Files: catalog.Files})
}

func (s *Server) readCatalog() (Catalog, error) {
	s.catalogMu.Lock()
	defer s.catalogMu.Unlock()

	return s.loadCatalogLocked()
}

func (s *Server) recordInitialVersion(root, sourcePath, displayName string, sizeBytes int64, minerIndex, uploadedAt string) (CatalogFile, CatalogVersion, error) {
	s.catalogMu.Lock()
	defer s.catalogMu.Unlock()

	catalog, err := s.loadCatalogLocked()
	if err != nil {
		return CatalogFile{}, CatalogVersion{}, err
	}

	file, version := catalog.upsertInitialVersion(root, sourcePath, displayName, sizeBytes, minerIndex, uploadedAt)
	if err := s.saveCatalogLocked(catalog); err != nil {
		return CatalogFile{}, CatalogVersion{}, err
	}

	return file, version, nil
}

func (s *Server) recordPatchVersion(cid, previousCID, sourcePath, displayName string, sizeBytes int64, minerIndex, uploadedAt string) (CatalogFile, CatalogVersion, error) {
	s.catalogMu.Lock()
	defer s.catalogMu.Unlock()

	catalog, err := s.loadCatalogLocked()
	if err != nil {
		return CatalogFile{}, CatalogVersion{}, err
	}

	file, version, err := catalog.appendPatchVersion(cid, previousCID, sourcePath, displayName, sizeBytes, minerIndex, uploadedAt)
	if err != nil {
		return CatalogFile{}, CatalogVersion{}, err
	}

	if err := s.saveCatalogLocked(catalog); err != nil {
		return CatalogFile{}, CatalogVersion{}, err
	}

	return file, version, nil
}

func (s *Server) loadCatalogLocked() (Catalog, error) {
	path := s.catalogPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Catalog{Files: []CatalogFile{}}, nil
		}
		return Catalog{}, fmt.Errorf("read catalog: %w", err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("parse catalog: %w", err)
	}

	catalog.normalize()
	return catalog, nil
}

func (s *Server) saveCatalogLocked(catalog Catalog) error {
	catalog.normalize()

	dir := filepath.Dir(s.catalogPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create catalog directory: %w", err)
	}

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal catalog: %w", err)
	}

	tmpPath := s.catalogPath() + ".tmp"
	if err := os.WriteFile(tmpPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("write catalog temp file: %w", err)
	}

	if err := os.Rename(tmpPath, s.catalogPath()); err != nil {
		return fmt.Errorf("replace catalog: %w", err)
	}

	return nil
}

func (s *Server) catalogPath() string {
	return filepath.Join(s.cfg.RepoRoot, catalogDirName, catalogFileName)
}

func (catalog *Catalog) upsertInitialVersion(root, sourcePath, displayName string, sizeBytes int64, minerIndex, uploadedAt string) (CatalogFile, CatalogVersion) {
	sourceName := deriveSourceName(sourcePath, displayName, root)
	extension := strings.ToLower(filepath.Ext(sourceName))

	if fileIndex := catalog.findFileIndexByRoot(root); fileIndex >= 0 {
		file := &catalog.Files[fileIndex]
		if strings.TrimSpace(file.DisplayName) == "" {
			file.DisplayName = sourceName
		}
		if strings.TrimSpace(file.Extension) == "" {
			file.Extension = extension
		}
		if strings.TrimSpace(file.CreatedAt) == "" {
			file.CreatedAt = uploadedAt
		}
		if versionIndex := findVersionIndex(file.Versions, root); versionIndex >= 0 {
			catalog.normalize()
			fileIndex = catalog.findFileIndexByRoot(root)
			versionIndex = findVersionIndex(catalog.Files[fileIndex].Versions, root)
			return catalog.Files[fileIndex], catalog.Files[fileIndex].Versions[versionIndex]
		}

		version := CatalogVersion{
			CID:                  root,
			RootCID:              root,
			VersionLabel:         "v1",
			PreviousCID:          "",
			PreviousVersionLabel: "",
			MinerIndex:           minerIndex,
			SourceName:           sourceName,
			SizeBytes:            sizeBytes,
			SizeMB:               formatSizeMB(sizeBytes),
			UploadedAt:           uploadedAt,
			Kind:                 "initial",
		}
		file.Versions = append(file.Versions, version)
		catalog.normalize()
		fileIndex = catalog.findFileIndexByRoot(root)
		versionIndex := findVersionIndex(catalog.Files[fileIndex].Versions, root)
		return catalog.Files[fileIndex], catalog.Files[fileIndex].Versions[versionIndex]
	}

	version := CatalogVersion{
		CID:                  root,
		RootCID:              root,
		VersionLabel:         "v1",
		PreviousCID:          "",
		PreviousVersionLabel: "",
		MinerIndex:           minerIndex,
		SourceName:           sourceName,
		SizeBytes:            sizeBytes,
		SizeMB:               formatSizeMB(sizeBytes),
		UploadedAt:           uploadedAt,
		Kind:                 "initial",
	}

	file := CatalogFile{
		RootCID:     root,
		DisplayName: sourceName,
		Extension:   extension,
		CreatedAt:   uploadedAt,
		Versions:    []CatalogVersion{version},
	}
	catalog.Files = append(catalog.Files, file)
	catalog.normalize()
	fileIndex := catalog.findFileIndexByRoot(root)
	return catalog.Files[fileIndex], catalog.Files[fileIndex].Versions[findVersionIndex(catalog.Files[fileIndex].Versions, root)]
}

func (catalog *Catalog) appendPatchVersion(cid, previousCID, sourcePath, displayName string, sizeBytes int64, minerIndex, uploadedAt string) (CatalogFile, CatalogVersion, error) {
	fileIndex, previousVersionIndex := catalog.findVersionLocation(previousCID)
	if fileIndex < 0 || previousVersionIndex < 0 {
		return CatalogFile{}, CatalogVersion{}, fmt.Errorf("previous version %s is not present in the catalog", previousCID)
	}

	file := &catalog.Files[fileIndex]
	if versionIndex := findVersionIndex(file.Versions, cid); versionIndex >= 0 {
		return *file, file.Versions[versionIndex], nil
	}

	previousVersion := file.Versions[previousVersionIndex]
	version := CatalogVersion{
		CID:                  cid,
		RootCID:              file.RootCID,
		VersionLabel:         "",
		PreviousCID:          previousVersion.CID,
		PreviousVersionLabel: "",
		MinerIndex:           minerIndex,
		SourceName:           deriveSourceName(sourcePath, displayName, cid),
		SizeBytes:            sizeBytes,
		SizeMB:               formatSizeMB(sizeBytes),
		UploadedAt:           uploadedAt,
		Kind:                 "patch",
	}

	file.Versions = append(file.Versions, version)
	catalog.normalize()

	fileIndex, versionIndex := catalog.findVersionLocation(cid)
	if fileIndex < 0 || versionIndex < 0 {
		return CatalogFile{}, CatalogVersion{}, fmt.Errorf("version %s was added but not found after normalization", cid)
	}

	return catalog.Files[fileIndex], catalog.Files[fileIndex].Versions[versionIndex], nil
}

func (catalog *Catalog) findFileIndexByRoot(root string) int {
	for i := range catalog.Files {
		if catalog.Files[i].RootCID == root {
			return i
		}
	}
	return -1
}

func (catalog *Catalog) findVersionLocation(cid string) (int, int) {
	for fileIndex := range catalog.Files {
		if versionIndex := findVersionIndex(catalog.Files[fileIndex].Versions, cid); versionIndex >= 0 {
			return fileIndex, versionIndex
		}
	}
	return -1, -1
}

func findVersionIndex(versions []CatalogVersion, cid string) int {
	for i := range versions {
		if versions[i].CID == cid {
			return i
		}
	}
	return -1
}

func (catalog *Catalog) normalize() {
	for i := range catalog.Files {
		catalog.Files[i].normalizeVersions()
	}

	sort.Slice(catalog.Files, func(i, j int) bool {
		left := strings.ToLower(catalog.Files[i].DisplayName)
		right := strings.ToLower(catalog.Files[j].DisplayName)
		if left != right {
			return left < right
		}
		return catalog.Files[i].CreatedAt < catalog.Files[j].CreatedAt
	})
}

func deriveSourceName(sourcePath, displayName, fallback string) string {
	candidate := strings.TrimSpace(displayName)
	if candidate != "" {
		return filepath.Base(candidate)
	}
	if strings.TrimSpace(sourcePath) != "" {
		return filepath.Base(sourcePath)
	}
	return fallback
}

func formatSizeMB(sizeBytes int64) string {
	if sizeBytes <= 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", float64(sizeBytes)/(1024*1024))
}

func (file *CatalogFile) normalizeVersions() {
	if len(file.Versions) == 0 {
		return
	}

	rootIndex := findVersionIndex(file.Versions, file.RootCID)
	if rootIndex < 0 {
		rootIndex = 0
	}

	children := make(map[string][]int)
	for index := range file.Versions {
		file.Versions[index].RootCID = file.RootCID
		if strings.TrimSpace(file.Versions[index].SourceName) == "" {
			file.Versions[index].SourceName = file.DisplayName
		}
		if file.Versions[index].Kind == "" {
			if file.Versions[index].CID == file.RootCID {
				file.Versions[index].Kind = "initial"
			} else {
				file.Versions[index].Kind = "patch"
			}
		}
		if file.Versions[index].SizeMB == "" {
			file.Versions[index].SizeMB = formatSizeMB(file.Versions[index].SizeBytes)
		}
		if previousCID := strings.TrimSpace(file.Versions[index].PreviousCID); previousCID != "" {
			children[previousCID] = append(children[previousCID], index)
		}
	}

	for parentCID := range children {
		sort.Slice(children[parentCID], func(i, j int) bool {
			left := file.Versions[children[parentCID][i]]
			right := file.Versions[children[parentCID][j]]
			if left.UploadedAt != right.UploadedAt {
				return left.UploadedAt < right.UploadedAt
			}
			if left.SourceName != right.SourceName {
				return left.SourceName < right.SourceName
			}
			return left.CID < right.CID
		})
	}

	labelByCID := make(map[string]string, len(file.Versions))
	orderedIndices := make([]int, 0, len(file.Versions))
	visited := make(map[string]struct{}, len(file.Versions))

	var walkTree func(versionIndex int, label string)
	walkTree = func(versionIndex int, label string) {
		cid := file.Versions[versionIndex].CID
		if _, ok := visited[cid]; ok {
			return
		}
		visited[cid] = struct{}{}
		labelByCID[cid] = label
		orderedIndices = append(orderedIndices, versionIndex)

		childIndices := children[cid]
		if len(childIndices) == 0 {
			return
		}

		continuationLabel := nextGenerationLabel(label)
		walkTree(childIndices[0], continuationLabel)
		for branchOffset, childIndex := range childIndices[1:] {
			branchLabel := fmt.Sprintf("%s.%d", continuationLabel, branchOffset+1)
			walkTree(childIndex, branchLabel)
		}
	}

	walkTree(rootIndex, "v1")

	if len(orderedIndices) < len(file.Versions) {
		orphanSeed := len(file.Versions) + 1
		orphanBlockSize := len(file.Versions) + 1
		orphanCount := 0
		for index := range file.Versions {
			if _, ok := visited[file.Versions[index].CID]; ok {
				continue
			}
			fallbackLabel := fmt.Sprintf("v%d", orphanSeed+orphanCount*orphanBlockSize)
			orphanCount++
			walkTree(index, fallbackLabel)
		}
	}

	orderedVersions := make([]CatalogVersion, 0, len(file.Versions))
	for _, index := range orderedIndices {
		version := file.Versions[index]
		version.VersionLabel = labelByCID[version.CID]
		if previousLabel, ok := labelByCID[version.PreviousCID]; ok {
			version.PreviousVersionLabel = previousLabel
		} else {
			version.PreviousVersionLabel = ""
		}
		orderedVersions = append(orderedVersions, version)
	}

	sort.SliceStable(orderedVersions, func(i, j int) bool {
		if compare := compareVersionLabels(orderedVersions[i].VersionLabel, orderedVersions[j].VersionLabel); compare != 0 {
			return compare < 0
		}
		if orderedVersions[i].UploadedAt != orderedVersions[j].UploadedAt {
			return orderedVersions[i].UploadedAt < orderedVersions[j].UploadedAt
		}
		return orderedVersions[i].CID < orderedVersions[j].CID
	})

	file.Versions = orderedVersions
}

func nextGenerationLabel(label string) string {
	parts, ok := parseVersionLabel(label)
	if !ok || len(parts) == 0 {
		return label + ".1"
	}
	parts[0]++
	return formatVersionLabel(parts)
}

func compareVersionLabels(left, right string) int {
	leftParts, leftOK := parseVersionLabel(left)
	rightParts, rightOK := parseVersionLabel(right)

	switch {
	case leftOK && rightOK:
		limit := len(leftParts)
		if len(rightParts) < limit {
			limit = len(rightParts)
		}
		for i := 0; i < limit; i++ {
			if leftParts[i] != rightParts[i] {
				if leftParts[i] < rightParts[i] {
					return -1
				}
				return 1
			}
		}
		switch {
		case len(leftParts) < len(rightParts):
			return -1
		case len(leftParts) > len(rightParts):
			return 1
		default:
			return 0
		}
	case leftOK:
		return -1
	case rightOK:
		return 1
	default:
		return strings.Compare(left, right)
	}
}

func parseVersionLabel(label string) ([]int, bool) {
	label = strings.TrimSpace(label)
	if len(label) < 2 || (label[0] != 'v' && label[0] != 'V') {
		return nil, false
	}

	rawParts := strings.Split(label[1:], ".")
	if len(rawParts) == 0 {
		return nil, false
	}

	parts := make([]int, 0, len(rawParts))
	for _, rawPart := range rawParts {
		if rawPart == "" {
			return nil, false
		}
		value, err := strconv.Atoi(rawPart)
		if err != nil {
			return nil, false
		}
		parts = append(parts, value)
	}

	return parts, true
}

func formatVersionLabel(parts []int) string {
	if len(parts) == 0 {
		return "v1"
	}

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, strconv.Itoa(part))
	}
	return "v" + strings.Join(values, ".")
}
