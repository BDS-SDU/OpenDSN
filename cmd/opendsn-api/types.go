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

type ImportRequest struct {
	Path string `json:"path"`
}

type ImportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Root    string `json:"root"`
	Output  string `json:"output"`
}

type DealRequest struct {
	Root         string `json:"root"`
	Miner        string `json:"miner"`
	PreviousRoot string `json:"previous_root,omitempty"`
}

type DealResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ProposalCID string `json:"proposal_cid,omitempty"`
	Output      string `json:"output"`
}

type RetrieveVersionRequest struct {
	Root       string `json:"root"`
	OutputName string `json:"output_name"`
}

type RetrieveVersionResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"output_path,omitempty"`
	Output     string `json:"output"`
}

type RootListResponse struct {
	Roots []RootInfo `json:"roots"`
}

type VersionListResponse struct {
	Root     string        `json:"root"`
	Versions []VersionInfo `json:"versions"`
}
