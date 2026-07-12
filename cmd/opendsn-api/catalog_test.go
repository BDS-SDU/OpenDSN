package main

import (
	"reflect"
	"testing"
)

func TestNormalizeVersions_ParallelBranchLabels(t *testing.T) {
	file := CatalogFile{
		RootCID:     "root",
		DisplayName: "report.docx",
		Versions: []CatalogVersion{
			{CID: "root", UploadedAt: "2026-07-11T10:00:00Z"},
			{CID: "v2cid", PreviousCID: "root", UploadedAt: "2026-07-11T10:01:00Z"},
			{CID: "v3cid", PreviousCID: "v2cid", UploadedAt: "2026-07-11T10:02:00Z"},
			{CID: "v4cid", PreviousCID: "v3cid", UploadedAt: "2026-07-11T10:03:00Z"},
			{CID: "v5cid", PreviousCID: "v4cid", UploadedAt: "2026-07-11T10:04:00Z"},
			{CID: "branchcid", PreviousCID: "v2cid", UploadedAt: "2026-07-11T10:05:00Z"},
		},
	}

	file.normalizeVersions()

	wantLabels := map[string]string{
		"root":      "v1",
		"v2cid":     "v2",
		"v3cid":     "v3",
		"v4cid":     "v4",
		"v5cid":     "v5",
		"branchcid": "v3.1",
	}

	gotLabels := labelsByCID(file.Versions)
	if !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Fatalf("unexpected labels:\nwant: %#v\ngot:  %#v", wantLabels, gotLabels)
	}

	branch := findVersionByCID(file.Versions, "branchcid")
	if branch == nil {
		t.Fatal("branch version not found after normalization")
	}
	if branch.PreviousVersionLabel != "v2" {
		t.Fatalf("unexpected previous version label for branch: want v2, got %s", branch.PreviousVersionLabel)
	}

	wantOrder := []string{"v1", "v2", "v3", "v3.1", "v4", "v5"}
	if gotOrder := orderedLabels(file.Versions); !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected version order:\nwant: %#v\ngot:  %#v", wantOrder, gotOrder)
	}
}

func TestNormalizeVersions_NestedBranchLabelsRemainUnique(t *testing.T) {
	file := CatalogFile{
		RootCID:     "root",
		DisplayName: "report.docx",
		Versions: []CatalogVersion{
			{CID: "root", UploadedAt: "2026-07-11T10:00:00Z"},
			{CID: "main2", PreviousCID: "root", UploadedAt: "2026-07-11T10:01:00Z"},
			{CID: "main3", PreviousCID: "main2", UploadedAt: "2026-07-11T10:02:00Z"},
			{CID: "main4", PreviousCID: "main3", UploadedAt: "2026-07-11T10:03:00Z"},
			{CID: "branch31", PreviousCID: "main2", UploadedAt: "2026-07-11T10:04:00Z"},
			{CID: "branch41", PreviousCID: "branch31", UploadedAt: "2026-07-11T10:05:00Z"},
			{CID: "branch51", PreviousCID: "branch41", UploadedAt: "2026-07-11T10:06:00Z"},
			{CID: "branch411", PreviousCID: "branch31", UploadedAt: "2026-07-11T10:07:00Z"},
			{CID: "branch32", PreviousCID: "main2", UploadedAt: "2026-07-11T10:08:00Z"},
		},
	}

	file.normalizeVersions()

	wantLabels := map[string]string{
		"root":      "v1",
		"main2":     "v2",
		"main3":     "v3",
		"main4":     "v4",
		"branch31":  "v3.1",
		"branch41":  "v4.1",
		"branch51":  "v5.1",
		"branch411": "v4.1.1",
		"branch32":  "v3.2",
	}

	gotLabels := labelsByCID(file.Versions)
	if !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Fatalf("unexpected labels:\nwant: %#v\ngot:  %#v", wantLabels, gotLabels)
	}

	wantOrder := []string{"v1", "v2", "v3", "v3.1", "v3.2", "v4", "v4.1", "v4.1.1", "v5.1"}
	if gotOrder := orderedLabels(file.Versions); !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("unexpected version order:\nwant: %#v\ngot:  %#v", wantOrder, gotOrder)
	}

	branch411 := findVersionByCID(file.Versions, "branch411")
	if branch411 == nil {
		t.Fatal("nested branch version not found after normalization")
	}
	if branch411.PreviousVersionLabel != "v3.1" {
		t.Fatalf("unexpected previous version label for nested branch: want v3.1, got %s", branch411.PreviousVersionLabel)
	}
}

func labelsByCID(versions []CatalogVersion) map[string]string {
	result := make(map[string]string, len(versions))
	for _, version := range versions {
		result[version.CID] = version.VersionLabel
	}
	return result
}

func orderedLabels(versions []CatalogVersion) []string {
	result := make([]string, 0, len(versions))
	for _, version := range versions {
		result = append(result, version.VersionLabel)
	}
	return result
}

func findVersionByCID(versions []CatalogVersion, cid string) *CatalogVersion {
	for i := range versions {
		if versions[i].CID == cid {
			return &versions[i]
		}
	}
	return nil
}
