package main

type Config struct {
	Addr          string
	RepoRoot      string
	LotusBin      string
	MinerHTTPPort int
	DealPrice     string
	DealDuration  string
}

type MinerInfo struct {
	NodeIP         string `json:"node_ip"`
	Index          string `json:"index"`
	StoragePower   string `json:"storage_power"`
	CommittedSpace string `json:"committed_space"`
	UserDataSize   string `json:"user_data_size"`
}

type ProofInfo struct {
	NodeIP                     string `json:"node_ip"`
	ProofType                  string `json:"proof_type"`
	Status                     string `json:"status"`
	Timestamp                  string `json:"timestamp"`
	GenerateDurationSeconds    string `json:"generate_duration_seconds"`
	VerifyDurationMilliseconds string `json:"verify_duration_milliseconds"`
}

type RootInfo struct {
	Root string `json:"root"`
}

type VersionInfo struct {
	Label string `json:"label"`
	CID   string `json:"cid"`
}

type BrowseEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

type BrowseResponse struct {
	Path    string        `json:"path"`
	Entries []BrowseEntry `json:"entries"`
}

type PreviewResponse struct {
	Filename    string `json:"filename"`
	Kind        string `json:"kind"`
	Content     string `json:"content,omitempty"`
	ResourceURL string `json:"resource_url,omitempty"`
}

type CatalogVersion struct {
	CID                  string `json:"cid"`
	RootCID              string `json:"root_cid"`
	VersionLabel         string `json:"version_label"`
	PreviousCID          string `json:"previous_cid,omitempty"`
	PreviousVersionLabel string `json:"previous_version_label,omitempty"`
	MinerIndex           string `json:"miner_index"`
	SourceName           string `json:"source_name"`
	SizeBytes            int64  `json:"size_bytes"`
	SizeMB               string `json:"size_mb"`
	UploadedAt           string `json:"uploaded_at"`
	Kind                 string `json:"kind"`
}

type CatalogFile struct {
	RootCID     string           `json:"root_cid"`
	DisplayName string           `json:"display_name"`
	Extension   string           `json:"extension"`
	CreatedAt   string           `json:"created_at"`
	Versions    []CatalogVersion `json:"versions"`
}

type CatalogResponse struct {
	Files []CatalogFile `json:"files"`
}

type ImportRequest struct {
	Path string `json:"path"`
}

type ImportResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Root          string `json:"root"`
	Output        string `json:"output"`
	Path          string `json:"path"`
	FileName      string `json:"file_name"`
	FileExtension string `json:"file_extension"`
	SizeBytes     int64  `json:"size_bytes"`
	SizeMB        string `json:"size_mb"`
}

type DealRequest struct {
	Root         string `json:"root"`
	Miner        string `json:"miner"`
	PreviousRoot string `json:"previous_root,omitempty"`
	SourcePath   string `json:"source_path,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	SizeBytes    int64  `json:"size_bytes,omitempty"`
}

type DealResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ProposalCID  string `json:"proposal_cid,omitempty"`
	Output       string `json:"output"`
	InitialRoot  string `json:"initial_root,omitempty"`
	VersionLabel string `json:"version_label,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	FileSizeMB   string `json:"file_size_mb,omitempty"`
	UploadedAt   string `json:"uploaded_at,omitempty"`
}

type RetrieveVersionRequest struct {
	Root       string `json:"root,omitempty"`
	CID        string `json:"cid,omitempty"`
	OutputName string `json:"output_name"`
}

type RetrieveVersionResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	OutputPath   string `json:"output_path,omitempty"`
	Output       string `json:"output"`
	RetrievedCID string `json:"retrieved_cid,omitempty"`
	SizeBytes    int64  `json:"size_bytes"`
	SizeMB       string `json:"size_mb"`
}

type RootListResponse struct {
	Roots []RootInfo `json:"roots"`
}

type VersionListResponse struct {
	Root     string        `json:"root"`
	Versions []VersionInfo `json:"versions"`
}
