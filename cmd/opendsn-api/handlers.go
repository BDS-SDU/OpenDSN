package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMiners(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	miners, err := s.collectMinerInfos()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"miners": miners})
}

func (s *Server) handleProofs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	proofs, err := s.collectProofInfos()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"proofs": proofs})
}

func (s *Server) handleClientImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	output, err := s.runLotusCommand(r.Context(), "client", "import", "-q", req.Path)
	root := strings.TrimSpace(output)

	if err != nil {
		writeJSON(w, http.StatusOK, ImportResponse{
			Success: false,
			Message: "import failed",
			Root:    root,
			Output:  output,
		})
		return
	}

	writeJSON(w, http.StatusOK, ImportResponse{
		Success: true,
		Message: "import finished",
		Root:    root,
		Output:  output,
	})
}

func (s *Server) handleClientDeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req DealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Root = strings.TrimSpace(req.Root)
	req.Miner = strings.TrimSpace(req.Miner)
	req.PreviousRoot = strings.TrimSpace(req.PreviousRoot)

	if req.Root == "" {
		writeError(w, http.StatusBadRequest, "root is required")
		return
	}
	if req.Miner == "" {
		writeError(w, http.StatusBadRequest, "miner is required")
		return
	}

	args := []string{"client", "deal"}
	if req.PreviousRoot == "" {
		args = append(args, "--create", req.Root, req.Miner, s.cfg.DealPrice, s.cfg.DealDuration)
	} else {
		args = append(args, "--update", "--previous", req.PreviousRoot, req.Root, req.Miner, s.cfg.DealPrice, s.cfg.DealDuration)
	}

	output, err := s.runLotusCommand(r.Context(), args...)
	proposalCID := strings.TrimSpace(output)
	if err != nil {
		writeJSON(w, http.StatusOK, DealResponse{
			Success:     false,
			Message:     "deal failed",
			ProposalCID: proposalCID,
			Output:      output,
		})
		return
	}

	writeJSON(w, http.StatusOK, DealResponse{
		Success:     true,
		Message:     "deal finished",
		ProposalCID: proposalCID,
		Output:      output,
	})
}

func (s *Server) handleClientRetrieveVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req RetrieveVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Root = strings.TrimSpace(req.Root)
	req.OutputName = strings.TrimSpace(req.OutputName)
	if req.Root == "" {
		writeError(w, http.StatusBadRequest, "root is required")
		return
	}
	if req.OutputName == "" {
		writeError(w, http.StatusBadRequest, "output_name is required")
		return
	}

	outputPath := filepath.Join(s.cfg.RepoRoot, req.OutputName)
	output, err := s.runLotusCommand(r.Context(), "client", "retrieve-version", req.Root, outputPath)
	if err != nil {
		writeJSON(w, http.StatusOK, RetrieveVersionResponse{
			Success:    false,
			Message:    "retrieve-version failed",
			OutputPath: outputPath,
			Output:     output,
		})
		return
	}

	writeJSON(w, http.StatusOK, RetrieveVersionResponse{
		Success:    true,
		Message:    "retrieve-version finished",
		OutputPath: outputPath,
		Output:     output,
	})
}

func (s *Server) handleClientRoots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	roots, err := loadRoots(filepath.Join(s.cfg.RepoRoot, "fileroots.log"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, RootListResponse{Roots: roots})
}

func (s *Server) handleClientVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	root := strings.TrimSpace(r.URL.Query().Get("root"))
	if root == "" {
		writeError(w, http.StatusBadRequest, "root is required")
		return
	}

	versions, err := loadVersionsFromMeta(s.cfg.RepoRoot, root)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, VersionListResponse{Root: root, Versions: versions})
}

func (s *Server) collectMinerInfos() ([]MinerInfo, error) {
	var out []MinerInfo

	localPath := filepath.Join(s.cfg.RepoRoot, "miner_info.log")
	localInfo, err := readMinerInfoFile(localPath)
	if err == nil {
		out = append(out, localInfo)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	localIP := localInfo.NodeIP
	ips, err := loadIPs(filepath.Join(s.cfg.RepoRoot, "miner_ip.log"))
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		if ip == "" {
			continue
		}
		if localIP != "" && ip == localIP {
			continue
		}
		info, err := s.fetchRemoteMinerInfo(ip)
		if err != nil {
			continue
		}
		out = append(out, info)
	}

	return out, nil
}

func (s *Server) collectProofInfos() ([]ProofInfo, error) {
	var out []ProofInfo

	localIP := ""
	localMinerInfo, err := readMinerInfoFile(filepath.Join(s.cfg.RepoRoot, "miner_info.log"))
	if err == nil {
		localIP = localMinerInfo.NodeIP
	}

	localProofs, err := readProofInfoFile(filepath.Join(s.cfg.RepoRoot, "proof_info.log"))
	if err == nil {
		for _, p := range localProofs {
			if strings.TrimSpace(p.Status) == "" {
				continue
			}
			p.NodeIP = localIP
			out = append(out, p)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	ips, err := loadIPs(filepath.Join(s.cfg.RepoRoot, "miner_ip.log"))
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		if ip == "" {
			continue
		}
		if localIP != "" && ip == localIP {
			continue
		}
		proofs, err := s.fetchRemoteProofInfo(ip)
		if err != nil {
			continue
		}
		for _, p := range proofs {
			if strings.TrimSpace(p.Status) == "" {
				continue
			}
			p.NodeIP = ip
			out = append(out, p)
		}
	}

	return out, nil
}

func (s *Server) fetchRemoteMinerInfo(ip string) (MinerInfo, error) {
	url := fmt.Sprintf("http://%s:%d/miner_info.log", ip, s.cfg.MinerHTTPPort)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return MinerInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MinerInfo{}, fmt.Errorf("http status %s", resp.Status)
	}

	info, err := parseMinerInfo(resp.Body)
	if err != nil {
		return MinerInfo{}, err
	}
	if info.NodeIP == "" {
		info.NodeIP = ip
	}
	return info, nil
}

func (s *Server) fetchRemoteProofInfo(ip string) ([]ProofInfo, error) {
	url := fmt.Sprintf("http://%s:%d/proof_info.log", ip, s.cfg.MinerHTTPPort)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %s", resp.Status)
	}

	return parseProofInfo(resp.Body)
}

func readMinerInfoFile(path string) (MinerInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return MinerInfo{}, err
	}
	defer f.Close()
	return parseMinerInfo(f)
}

func parseMinerInfo(r io.Reader) (MinerInfo, error) {
	var info MinerInfo

	scanner := bufio.NewScanner(r)
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
		case "ip":
			info.NodeIP = val
		case "miner":
			info.Index = val
		case "quality_adjusted_power":
			info.StoragePower = val
		case "committed_space":
			info.CommittedSpace = val
		case "user_data_size":
			info.UserDataSize = val
		}
	}
	if err := scanner.Err(); err != nil {
		return MinerInfo{}, err
	}
	return info, nil
}

func readProofInfoFile(path string) ([]ProofInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseProofInfo(f)
}

func parseProofInfo(r io.Reader) ([]ProofInfo, error) {
	records := map[string]*ProofInfo{
		"storage_proof": {
			ProofType:                  "storage_proof",
			GenerateDurationSeconds:    "0",
			VerifyDurationMilliseconds: "0",
		},
		"window_post": {
			ProofType:                  "window_post",
			GenerateDurationSeconds:    "0",
			VerifyDurationMilliseconds: "0",
		},
	}

	scanner := bufio.NewScanner(r)
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
			records["storage_proof"].ProofType = val
		case "storage_proof_status":
			records["storage_proof"].Status = val
		case "storage_proof_timestamp":
			records["storage_proof"].Timestamp = val
		case "storage_proof_generate_duration_seconds":
			records["storage_proof"].GenerateDurationSeconds = val
		case "storage_proof_verify_duration_milliseconds":
			records["storage_proof"].VerifyDurationMilliseconds = val
		case "storage_proof_verify_duration_seconds":
			records["storage_proof"].VerifyDurationMilliseconds = val
		case "window_post_type":
			records["window_post"].ProofType = val
		case "window_post_status":
			records["window_post"].Status = val
		case "window_post_timestamp":
			records["window_post"].Timestamp = val
		case "window_post_generate_duration_seconds":
			records["window_post"].GenerateDurationSeconds = val
		case "window_post_verify_duration_milliseconds":
			records["window_post"].VerifyDurationMilliseconds = val
		case "window_post_verify_duration_seconds":
			records["window_post"].VerifyDurationMilliseconds = val
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []ProofInfo{*records["storage_proof"], *records["window_post"]}, nil
}

func loadIPs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var ordered []string
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip == "" {
			continue
		}
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		ordered = append(ordered, ip)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ordered, nil
}

func loadRoots(path string) ([]RootInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var roots []RootInfo
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		root := strings.TrimSpace(scanner.Text())
		if root == "" {
			continue
		}
		if _, ok := seen[root]; ok {
			continue
		}
		seen[root] = struct{}{}
		roots = append(roots, RootInfo{Root: root})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return roots, nil
}

func loadVersionsFromMeta(repoRoot, root string) ([]VersionInfo, error) {
	metaPath := filepath.Join(repoRoot, fmt.Sprintf("%s_meta", root))
	if _, err := os.Stat(metaPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metadata file not found for root %s", root)
		}
		return nil, err
	}

	visited := make(map[string]struct{})
	current := root
	index := 1
	versions := []VersionInfo{{Label: "ROOT(v1)", CID: root}}

	for {
		if _, ok := visited[current]; ok {
			return nil, fmt.Errorf("detected cycle in metadata at %s", current)
		}
		visited[current] = struct{}{}

		content, err := os.ReadFile(filepath.Join(repoRoot, fmt.Sprintf("%s_meta", current)))
		if err != nil {
			return nil, err
		}

		normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
		normalized = strings.TrimRight(normalized, "\n")
		lines := strings.Split(normalized, "\n")
		if len(lines) < 2 {
			return nil, fmt.Errorf("invalid metadata for %s", current)
		}

		nextCID := strings.TrimSpace(lines[1])
		if nextCID == "" {
			return nil, fmt.Errorf("invalid metadata for %s: empty head", current)
		}
		if nextCID == "NULL" {
			break
		}

		index++
		versions = append(versions, VersionInfo{Label: fmt.Sprintf("HEAD(v%d)", index), CID: nextCID})
		current = nextCID
	}

	return versions, nil
}

func (s *Server) runLotusCommand(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.cfg.LotusBin, args...)
	cmd.Dir = s.cfg.RepoRoot
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
