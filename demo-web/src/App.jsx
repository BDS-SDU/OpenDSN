import React, { useEffect, useMemo, useState } from "react";
import sduLogo from "../sdu_logo.png";
import opendsnLogo from "../opendsn_logo.png";

const API_BASE = (import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080").replace(/\/$/, "");
const PAGE_SIZE = 5;

function apiURL(path) {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  if (API_BASE.endsWith("/api")) {
    return `${API_BASE}${normalized.replace(/^\/api/, "")}`;
  }
  return `${API_BASE}${normalized}`;
}

function formatTimestamp(value) {
  if (!value) {
    return "-";
  }
  const match = String(value).match(/^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})/);
  if (match) {
    return `${match[1]} ${match[2]}`;
  }
  return String(value).replace("T", " ").replace(/\+\d{2}:\d{2}$/, "");
}

function formatSizeWithUnit(sizeMB) {
  if (!sizeMB) {
    return "-";
  }
  return `${sizeMB} MB`;
}

function getFileExtension(fileName) {
  const normalized = String(fileName || "").split("/").pop() || "";
  const lastDot = normalized.lastIndexOf(".");
  if (lastDot <= 0 || lastDot === normalized.length - 1) {
    return "";
  }
  return normalized.slice(lastDot);
}

function stripTrailingExtension(value, extension) {
  if (!extension) {
    return value;
  }
  return value.endsWith(extension) ? value.slice(0, -extension.length) : value;
}

function paginate(items, page, pageSize) {
  const totalPages = Math.max(1, Math.ceil(items.length / pageSize));
  const safePage = Math.min(Math.max(1, page), totalPages);
  const start = (safePage - 1) * pageSize;
  return {
    pageItems: items.slice(start, start + pageSize),
    totalPages,
    safePage,
  };
}

async function getJSON(path) {
  const res = await fetch(apiURL(path));
  if (!res.ok) {
    throw new Error(`Request failed: ${res.status}`);
  }
  return res.json();
}

async function postJSON(path, body) {
  const res = await fetch(apiURL(path), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    throw new Error(`Request failed: ${res.status}`);
  }

  return res.json();
}

function buildFileLabel(file, duplicateCounts) {
  if (!file) {
    return "";
  }
  if ((duplicateCounts[file.display_name] || 0) <= 1) {
    return file.display_name;
  }
  return `${file.display_name} (${formatTimestamp(file.created_at)})`;
}

function SectionTitle({ children }) {
  return <h2 className="section-title">{children}</h2>;
}

function Pager({ page, totalPages, onPrev, onNext }) {
  return (
    <div className="pager">
      <button onClick={onPrev} disabled={page <= 1}>
        Prev
      </button>
      <span>
        {page} / {totalPages}
      </span>
      <button onClick={onNext} disabled={page >= totalPages}>
        Next
      </button>
    </div>
  );
}

function FileBrowserModal({ onClose, onSelect }) {
  const [currentDir, setCurrentDir] = useState("/");
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;

    async function loadDir(dir) {
      try {
        setLoading(true);
        setError("");
        const data = await getJSON(`/api/browse?dir=${encodeURIComponent(dir)}`);
        if (cancelled) {
          return;
        }
        setCurrentDir(data.path || dir);
        setEntries(data.entries || []);
      } catch (err) {
        if (!cancelled) {
          setError(err.message);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    loadDir("/");
    return () => {
      cancelled = true;
    };
  }, []);

  async function navigateTo(dir) {
    try {
      setLoading(true);
      setError("");
      const data = await getJSON(`/api/browse?dir=${encodeURIComponent(dir)}`);
      setCurrentDir(data.path || dir);
      setEntries(data.entries || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function handleEntryClick(entry) {
    if (entry.is_dir) {
      const nextPath = currentDir === "/" ? `/${entry.name}` : `${currentDir}/${entry.name}`;
      navigateTo(nextPath);
      return;
    }

    const fullPath = currentDir === "/" ? `/${entry.name}` : `${currentDir}/${entry.name}`;
    onSelect(fullPath);
    onClose();
  }

  function handleParent() {
    const parts = currentDir.split("/").filter(Boolean);
    parts.pop();
    const parent = parts.length === 0 ? "/" : `/${parts.join("/")}`;
    navigateTo(parent);
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content file-browser-modal" onClick={(event) => event.stopPropagation()}>
        <div className="modal-header">
          <h3>Select File</h3>
          <button className="modal-close" onClick={onClose}>
            ✕
          </button>
        </div>

        <div className="modal-body">
          <div className="browser-path">
            <button className="browser-parent-btn" onClick={handleParent}>
              ↑ Up
            </button>
            <span className="browser-current-path">{currentDir}</span>
          </div>

          {loading ? (
            <div className="empty-state">Loading...</div>
          ) : error ? (
            <div className="empty-state">{error}</div>
          ) : entries.length === 0 ? (
            <div className="empty-state">No files in this directory.</div>
          ) : (
            <div className="browser-list">
              {entries.map((entry, index) => (
                <div
                  key={`${entry.name}-${index}`}
                  className={`browser-entry ${entry.is_dir ? "is-dir" : "is-file"}`}
                  onClick={() => handleEntryClick(entry)}
                >
                  <span className="browser-entry-icon">{entry.is_dir ? "📁" : "📄"}</span>
                  <span className="browser-entry-name">{entry.name}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function PreviewModal({ title, mode, content, resourceUrl, onClose }) {
  const isPDF = mode === "pdf" && !!resourceUrl;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className={`modal-content preview-modal ${isPDF ? "is-pdf" : ""}`} onClick={(event) => event.stopPropagation()}>
        <div className="modal-header">
          <h3>{title || "Preview"}</h3>
          <button className="modal-close" onClick={onClose}>
            ✕
          </button>
        </div>

        <div className={`modal-body preview-body ${isPDF ? "pdf-preview-body" : ""}`}>
          {isPDF ? (
            <iframe className="preview-frame" title={title || "Preview"} src={apiURL(resourceUrl)} />
          ) : (
            <pre className="text-preview">{content || ""}</pre>
          )}
        </div>
      </div>
    </div>
  );
}

export default function App() {
  const [miners, setMiners] = useState([]);
  const [proofs, setProofs] = useState([]);
  const [catalogFiles, setCatalogFiles] = useState([]);

  const [minersLoading, setMinersLoading] = useState(false);
  const [proofsLoading, setProofsLoading] = useState(false);
  const [catalogLoading, setCatalogLoading] = useState(false);
  const [catalogError, setCatalogError] = useState("");

  const [minerPage, setMinerPage] = useState(1);
  const [proofPage, setProofPage] = useState(1);

  const [uploadMode, setUploadMode] = useState("initial");
  const [showFileBrowser, setShowFileBrowser] = useState(false);
  const [uploadFilePath, setUploadFilePath] = useState("");
  const [uploadImportInfo, setUploadImportInfo] = useState(null);
  const [uploadImporting, setUploadImporting] = useState(false);
  const [uploadLoading, setUploadLoading] = useState(false);
  const [uploadMessage, setUploadMessage] = useState("Ready.");
  const [uploadInitialFileRoot, setUploadInitialFileRoot] = useState("");
  const [uploadPreviousVersionCID, setUploadPreviousVersionCID] = useState("");
  const [uploadMinerIndex, setUploadMinerIndex] = useState("");

  const [retrieveFileRoot, setRetrieveFileRoot] = useState("");
  const [retrieveVersionCID, setRetrieveVersionCID] = useState("");
  const [retrieveOutputBaseName, setRetrieveOutputBaseName] = useState("");
  const [retrieveLoading, setRetrieveLoading] = useState(false);
  const [retrieveMessage, setRetrieveMessage] = useState("Ready.");
  const [retrieveSuccess, setRetrieveSuccess] = useState(false);
  const [retrieveOutputPath, setRetrieveOutputPath] = useState("");

  const [informationSelectionRoot, setInformationSelectionRoot] = useState("");
  const [informationActiveRoot, setInformationActiveRoot] = useState("");
  const [informationMessage, setInformationMessage] = useState("Ready.");

  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewTitle, setPreviewTitle] = useState("");
  const [previewMode, setPreviewMode] = useState("text");
  const [previewContent, setPreviewContent] = useState("");
  const [previewResourceUrl, setPreviewResourceUrl] = useState("");
  const [previewLoading, setPreviewLoading] = useState(false);

  const fileNameCounts = useMemo(() => {
    const counts = {};
    catalogFiles.forEach((file) => {
      counts[file.display_name] = (counts[file.display_name] || 0) + 1;
    });
    return counts;
  }, [catalogFiles]);

  const minerOptions = useMemo(
    () => Array.from(new Set(miners.map((item) => item.index).filter(Boolean))).sort(),
    [miners]
  );

  const uploadTargetFile = useMemo(
    () => catalogFiles.find((file) => file.root_cid === uploadInitialFileRoot) || null,
    [catalogFiles, uploadInitialFileRoot]
  );

  const uploadPreviousVersion = useMemo(
    () => uploadTargetFile?.versions.find((version) => version.cid === uploadPreviousVersionCID) || null,
    [uploadTargetFile, uploadPreviousVersionCID]
  );

  const retrieveTargetFile = useMemo(
    () => catalogFiles.find((file) => file.root_cid === retrieveFileRoot) || null,
    [catalogFiles, retrieveFileRoot]
  );

  const retrieveTargetVersion = useMemo(
    () => retrieveTargetFile?.versions.find((version) => version.cid === retrieveVersionCID) || null,
    [retrieveTargetFile, retrieveVersionCID]
  );

  const retrieveOutputExtension = useMemo(
    () => retrieveTargetFile?.extension || getFileExtension(retrieveTargetFile?.display_name || ""),
    [retrieveTargetFile]
  );

  const activeInformationFile = useMemo(
    () => catalogFiles.find((file) => file.root_cid === informationActiveRoot) || null,
    [catalogFiles, informationActiveRoot]
  );

  const minerPagination = useMemo(
    () => paginate(miners, minerPage, PAGE_SIZE),
    [miners, minerPage]
  );

  const proofPagination = useMemo(
    () => paginate(proofs, proofPage, PAGE_SIZE),
    [proofs, proofPage]
  );

  async function fetchMiners() {
    try {
      setMinersLoading(true);
      const data = await getJSON("/api/miners");
      setMiners(data.miners || []);
    } catch (err) {
      console.error(err);
    } finally {
      setMinersLoading(false);
    }
  }

  async function fetchProofs() {
    try {
      setProofsLoading(true);
      const data = await getJSON("/api/proofs");
      setProofs(data.proofs || []);
    } catch (err) {
      console.error(err);
    } finally {
      setProofsLoading(false);
    }
  }

  async function fetchCatalog(options = {}) {
    const { preserveOnError = true } = options;
    try {
      setCatalogLoading(true);
      const data = await getJSON("/api/catalog");
      const files = data.files || [];
      setCatalogFiles(files);
      setCatalogError("");
      return files;
    } catch (err) {
      if (!preserveOnError) {
        setCatalogFiles([]);
      }
      setCatalogError(err.message);
      return null;
    } finally {
      setCatalogLoading(false);
    }
  }

  useEffect(() => {
    fetchMiners();
    fetchProofs();
    fetchCatalog({ preserveOnError: false });

    const timer = setInterval(() => {
      fetchMiners();
      fetchProofs();
    }, 30000);

    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (minerPagination.safePage !== minerPage) {
      setMinerPage(minerPagination.safePage);
    }
  }, [minerPage, minerPagination.safePage]);

  useEffect(() => {
    if (proofPagination.safePage !== proofPage) {
      setProofPage(proofPagination.safePage);
    }
  }, [proofPage, proofPagination.safePage]);

  useEffect(() => {
    if (uploadMode === "initial") {
      setUploadInitialFileRoot("");
      setUploadPreviousVersionCID("");
    }
  }, [uploadMode]);

  useEffect(() => {
    setUploadPreviousVersionCID("");
  }, [uploadInitialFileRoot]);

  useEffect(() => {
    setRetrieveVersionCID("");
    setRetrieveOutputBaseName("");
    setRetrieveSuccess(false);
    setRetrieveOutputPath("");
    if (retrieveMessage !== "Ready.") {
      setRetrieveMessage("Ready.");
    }
  }, [retrieveFileRoot]);

  useEffect(() => {
    if (uploadInitialFileRoot && !catalogFiles.some((file) => file.root_cid === uploadInitialFileRoot)) {
      setUploadInitialFileRoot("");
      setUploadPreviousVersionCID("");
    }
    if (retrieveFileRoot && !catalogFiles.some((file) => file.root_cid === retrieveFileRoot)) {
      setRetrieveFileRoot("");
      setRetrieveVersionCID("");
    }
    if (informationSelectionRoot && !catalogFiles.some((file) => file.root_cid === informationSelectionRoot)) {
      setInformationSelectionRoot("");
    }
    if (informationActiveRoot && !catalogFiles.some((file) => file.root_cid === informationActiveRoot)) {
      setInformationActiveRoot("");
      setInformationMessage("Ready.");
    }
  }, [catalogFiles, uploadInitialFileRoot, retrieveFileRoot, informationSelectionRoot, informationActiveRoot]);

  async function openPreviewForFile(filePath) {
    try {
      setPreviewLoading(true);
      const data = await getJSON(`/api/preview?file=${encodeURIComponent(filePath)}`);
      setPreviewTitle(data.filename || "Preview");
      setPreviewMode(data.kind || "text");
      setPreviewContent(data.content || "");
      setPreviewResourceUrl(data.resource_url || "");
      setPreviewOpen(true);
    } catch (err) {
      throw err;
    } finally {
      setPreviewLoading(false);
    }
  }

  async function handleUploadFileSelect(fullPath) {
    setUploadFilePath(fullPath);
    setUploadImportInfo(null);
    setUploadMessage("Importing selected file...");

    try {
      setUploadImporting(true);
      const data = await postJSON("/api/client/import", { path: fullPath });
      if (!data.success) {
        setUploadMessage(data.output ? `Import failed.\n${data.output}` : "Import failed.");
        return;
      }

      setUploadImportInfo(data);
      setUploadMessage(
        [
          "Import finished.",
          `File: ${data.file_name || "-"}`,
          `Root: ${data.root || "-"}`,
          `Selected file size: ${formatSizeWithUnit(data.size_mb)}`,
        ].join("\n")
      );
    } catch (err) {
      setUploadMessage(`Request failed: ${err.message}`);
    } finally {
      setUploadImporting(false);
    }
  }

  async function handleUploadPreview() {
    if (uploadMode !== "initial" || !uploadFilePath.trim()) {
      return;
    }

    try {
      await openPreviewForFile(uploadFilePath.trim());
    } catch (err) {
      setUploadMessage(`Preview failed: ${err.message}`);
    }
  }

  async function handleUpload() {
    if (!uploadImportInfo?.root) {
      setUploadMessage("Please browse and select a file first.");
      return;
    }
    if (!uploadMinerIndex) {
      setUploadMessage("Please select a miner index.");
      return;
    }
    if (uploadMode === "patch") {
      if (!uploadInitialFileRoot) {
        setUploadMessage("Please select the initial file for this patch.");
        return;
      }
      if (!uploadPreviousVersion) {
        setUploadMessage("Please select the previous version for this patch.");
        return;
      }
    }

    try {
      setUploadLoading(true);
      setUploadMessage("Uploading...");

      const body = {
        root: uploadImportInfo.root,
        miner: uploadMinerIndex,
        source_path: uploadImportInfo.path,
        display_name: uploadImportInfo.file_name,
        size_bytes: uploadImportInfo.size_bytes,
      };

      if (uploadMode === "patch") {
        body.previous_root = uploadPreviousVersion.cid;
      }

      const data = await postJSON("/api/client/deal", body);
      if (!data.success) {
        setUploadMessage(data.output ? `Deal failed.\n${data.output}` : "Deal failed.");
        return;
      }

      const lines = ["Deal finished."];
      if (data.version_label) {
        lines.push(`Version: ${data.version_label}`);
      }
      if (uploadImportInfo.size_mb) {
        lines.push(`Selected file size: ${formatSizeWithUnit(uploadImportInfo.size_mb)}`);
      }
      if (data.message && data.message !== "deal finished") {
        lines.push(data.message);
      }
      setUploadMessage(lines.join("\n"));

      await Promise.all([fetchCatalog(), fetchMiners(), fetchProofs()]);
    } catch (err) {
      setUploadMessage(`Request failed: ${err.message}`);
    } finally {
      setUploadLoading(false);
    }
  }

  async function handleRetrieve() {
    if (!retrieveTargetFile) {
      setRetrieveMessage("Please select a file.");
      return;
    }
    if (!retrieveTargetVersion) {
      setRetrieveMessage("Please select a version.");
      return;
    }
    if (!retrieveOutputBaseName.trim()) {
      setRetrieveMessage("Please enter an output file name.");
      return;
    }

    try {
      setRetrieveLoading(true);
      setRetrieveSuccess(false);
      setRetrieveOutputPath("");
      setRetrieveMessage("Retrieving...");

      const data = await postJSON("/api/client/retrieve-version", {
        cid: retrieveTargetVersion.cid,
        output_name: `${retrieveOutputBaseName.trim()}${retrieveOutputExtension}`,
      });

      if (!data.success) {
        setRetrieveMessage(data.output ? `Retrieve failed.\n${data.output}` : "Retrieve failed.");
        return;
      }

      setRetrieveSuccess(true);
      setRetrieveOutputPath(data.output_path || "");
      setRetrieveMessage(
        [
          "Retrieve finished.",
          data.output_path ? `Output path: ${data.output_path}` : "",
          data.size_mb ? `Final file size: ${formatSizeWithUnit(data.size_mb)}` : "",
        ]
          .filter(Boolean)
          .join("\n")
      );
    } catch (err) {
      setRetrieveMessage(`Request failed: ${err.message}`);
    } finally {
      setRetrieveLoading(false);
    }
  }

  async function handleRetrievePreview() {
    if (!retrieveOutputPath) {
      return;
    }

    try {
      await openPreviewForFile(retrieveOutputPath);
    } catch (err) {
      setRetrieveMessage(`Preview failed: ${err.message}`);
    }
  }

  async function handleInformationConfirm() {
    if (!informationSelectionRoot) {
      setInformationActiveRoot("");
      setInformationMessage("Please select an initial file.");
      return;
    }

    const selectedRoot = informationSelectionRoot;
    const files = await fetchCatalog();
    if (!files) {
      setInformationMessage("Failed to refresh file catalog.");
      return;
    }
    if (!files.some((file) => file.root_cid === selectedRoot)) {
      setInformationActiveRoot("");
      setInformationMessage("Selected initial file is no longer available.");
      return;
    }

    setInformationActiveRoot(selectedRoot);
    setInformationMessage("Ready.");
  }

  return (
    <div className="app-shell">
      <header className="hero">
        <div className="hero-inner">
          <div className="hero-brand">
            <div className="hero-logo-frame hero-logo-frame-left">
              <img className="hero-logo hero-logo-sdu" src={sduLogo} alt="Shandong University logo" />
            </div>

            <div className="hero-title-block">
              <h1 className="hero-title">数链融合技术教育部工程研究中心——OpenDSN平台</h1>
            </div>

            <div className="hero-logo-frame hero-logo-frame-right">
              <img className="hero-logo hero-logo-opendsn" src={opendsnLogo} alt="OpenDSN logo" />
            </div>
          </div>

          <div className="hero-divider" />
        </div>
      </header>

      <main className="content">
        <section className="info-section">
          <div className="panel">
            <SectionTitle>Storage Node Information</SectionTitle>
            <div className="table-wrap">
              {minersLoading ? (
                <div className="empty-state">Refreshing storage node information...</div>
              ) : minerPagination.pageItems.length === 0 ? (
                <div className="empty-state">No storage node information available.</div>
              ) : (
                <table className="info-table">
                  <thead>
                    <tr>
                      <th>Node IP</th>
                      <th>Index</th>
                      <th>Storage Power</th>
                      <th>Committed Space</th>
                      <th>User Data Size</th>
                    </tr>
                  </thead>
                  <tbody>
                    {minerPagination.pageItems.map((item, idx) => (
                      <tr key={`${item.node_ip}-${item.index}-${idx}`}>
                        <td>{item.node_ip}</td>
                        <td>{item.index}</td>
                        <td>{item.storage_power}</td>
                        <td>{item.committed_space}</td>
                        <td>{item.user_data_size}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            <Pager
              page={minerPagination.safePage}
              totalPages={minerPagination.totalPages}
              onPrev={() => setMinerPage((page) => Math.max(1, page - 1))}
              onNext={() => setMinerPage((page) => Math.min(minerPagination.totalPages, page + 1))}
            />
          </div>

          <div className="panel">
            <SectionTitle>Proof Information</SectionTitle>
            <div className="table-wrap">
              {proofsLoading ? (
                <div className="empty-state">Refreshing proof information...</div>
              ) : proofPagination.pageItems.length === 0 ? (
                <div className="empty-state">No proof information available.</div>
              ) : (
                <table className="info-table proof-table">
                  <thead>
                    <tr>
                      <th>Node IP</th>
                      <th>Proof Type</th>
                      <th>Status</th>
                      <th>Timestamp</th>
                      <th>Generate (s)</th>
                      <th>Verify (ms)</th>
                    </tr>
                  </thead>
                  <tbody>
                    {proofPagination.pageItems.map((item, idx) => (
                      <tr key={`${item.node_ip}-${item.proof_type}-${idx}`}>
                        <td>{item.node_ip}</td>
                        <td>{item.proof_type}</td>
                        <td>{item.status}</td>
                        <td>{formatTimestamp(item.timestamp)}</td>
                        <td>{item.generate_duration_seconds}</td>
                        <td>{item.verify_duration_milliseconds}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            <Pager
              page={proofPagination.safePage}
              totalPages={proofPagination.totalPages}
              onPrev={() => setProofPage((page) => Math.max(1, page - 1))}
              onNext={() => setProofPage((page) => Math.min(proofPagination.totalPages, page + 1))}
            />
          </div>
        </section>

        <section className="workflow-row">
          <div className="card workflow-card">
            <SectionTitle>File Upload</SectionTitle>

            <div className="mode-toggle">
              <label className="radio-pill">
                <input
                  type="radio"
                  name="upload-mode"
                  value="initial"
                  checked={uploadMode === "initial"}
                  onChange={() => setUploadMode("initial")}
                />
                <span>Upload Initial Version</span>
              </label>

              <label className="radio-pill">
                <input
                  type="radio"
                  name="upload-mode"
                  value="patch"
                  checked={uploadMode === "patch"}
                  onChange={() => setUploadMode("patch")}
                />
                <span>Upload New Version</span>
              </label>
            </div>

            <label className="field-label">Selected File</label>
            <div className="file-input-row">
              <button className="browse-btn" onClick={() => setShowFileBrowser(true)} disabled={uploadImporting || uploadLoading}>
                Browse
              </button>
              <input
                type="text"
                className="file-path-input"
                placeholder="Select a file..."
                value={uploadFilePath}
                readOnly
              />
            </div>

            <button
              className="preview-btn"
              onClick={handleUploadPreview}
              disabled={uploadMode !== "initial" || !uploadFilePath.trim() || previewLoading}
            >
              {previewLoading ? "Loading..." : "Preview"}
            </button>

            <label className="field-label">File Root</label>
            <input type="text" value={uploadImportInfo?.root || ""} readOnly placeholder="Automatically filled after import" />

            <label className="field-label">Previous Version Selection</label>
            <div className={`previous-version-block ${uploadMode === "initial" ? "is-disabled" : ""}`}>
              <div className="dual-field-row">
                <select
                  value={uploadInitialFileRoot}
                  onChange={(event) => setUploadInitialFileRoot(event.target.value)}
                  disabled={uploadMode === "initial" || catalogFiles.length === 0}
                >
                  <option value="">Please select an initial file</option>
                  {catalogFiles.map((file) => (
                    <option key={file.root_cid} value={file.root_cid}>
                      {buildFileLabel(file, fileNameCounts)}
                    </option>
                  ))}
                </select>

                <select
                  value={uploadPreviousVersionCID}
                  onChange={(event) => setUploadPreviousVersionCID(event.target.value)}
                  disabled={uploadMode === "initial" || !uploadTargetFile}
                >
                  <option value="">Please select a version</option>
                  {(uploadTargetFile?.versions || []).map((version) => (
                    <option key={version.cid} value={version.cid}>
                      {version.version_label}
                    </option>
                  ))}
                </select>
              </div>

              <label className="field-label">Previous Version Hash</label>
              <input
                type="text"
                value={uploadPreviousVersion?.cid || ""}
                readOnly
                disabled={uploadMode === "initial"}
                placeholder="Automatically filled after version selection"
              />
            </div>

            <label className="field-label">Miner Index</label>
            <select value={uploadMinerIndex} onChange={(event) => setUploadMinerIndex(event.target.value)} disabled={minerOptions.length === 0}>
              <option value="">Please select a miner</option>
              {minerOptions.map((minerIndex) => (
                <option key={minerIndex} value={minerIndex}>
                  {minerIndex}
                </option>
              ))}
            </select>

            <button onClick={handleUpload} disabled={uploadImporting || uploadLoading}>
              {uploadImporting ? "Importing..." : uploadLoading ? "Uploading..." : "Upload File"}
            </button>

            <div className="message-box tall-message-box">
              <div className={uploadMessage === "Ready." ? "empty-state" : ""}>{uploadMessage}</div>
            </div>
          </div>

          <div className="card workflow-card">
            <SectionTitle>File Retrieval</SectionTitle>

            <label className="field-label">Retrieve Target</label>
            <div className="dual-field-row">
              <select
                value={retrieveFileRoot}
                onChange={(event) => setRetrieveFileRoot(event.target.value)}
                disabled={catalogFiles.length === 0}
              >
                <option value="">Please select a file</option>
                {catalogFiles.map((file) => (
                  <option key={file.root_cid} value={file.root_cid}>
                    {buildFileLabel(file, fileNameCounts)}
                  </option>
                ))}
              </select>

              <select
                value={retrieveVersionCID}
                onChange={(event) => setRetrieveVersionCID(event.target.value)}
                disabled={!retrieveTargetFile}
              >
                <option value="">Please select a version</option>
                {(retrieveTargetFile?.versions || []).map((version) => (
                  <option key={version.cid} value={version.cid}>
                    {version.version_label}
                  </option>
                ))}
              </select>
            </div>

            <label className="field-label">Selected File Hash</label>
            <input
              type="text"
              value={retrieveTargetVersion?.cid || ""}
              readOnly
              placeholder="Automatically filled after version selection"
            />

            <label className="field-label">Output File Name</label>
            <div className={`output-name-group ${retrieveOutputExtension ? "has-suffix" : ""}`}>
              <input
                type="text"
                placeholder="Enter output name"
                value={retrieveOutputBaseName}
                onChange={(event) =>
                  setRetrieveOutputBaseName(
                    stripTrailingExtension(event.target.value, retrieveOutputExtension)
                  )
                }
                disabled={retrieveLoading}
              />
              {retrieveOutputExtension ? <span className="output-name-suffix">{retrieveOutputExtension}</span> : null}
            </div>

            <button onClick={handleRetrieve} disabled={retrieveLoading}>
              {retrieveLoading ? "Retrieving..." : "Retrieve File"}
            </button>

            <button className="preview-btn" onClick={handleRetrievePreview} disabled={!retrieveSuccess || previewLoading}>
              {previewLoading ? "Loading..." : "Preview"}
            </button>

            <div className="message-box tall-message-box">
              <div className={retrieveMessage === "Ready." ? "empty-state" : ""}>{retrieveMessage}</div>
            </div>
          </div>
        </section>

        <section className="catalog-row">
          <div className="card catalog-card">
            <SectionTitle>File Information</SectionTitle>

            <div className="catalog-toolbar">
              <select
                value={informationSelectionRoot}
                onChange={(event) => setInformationSelectionRoot(event.target.value)}
                disabled={catalogFiles.length === 0}
              >
                <option value="">Please select an initial file</option>
                {catalogFiles.map((file) => (
                  <option key={file.root_cid} value={file.root_cid}>
                    {buildFileLabel(file, fileNameCounts)}
                  </option>
                ))}
              </select>

              <button onClick={handleInformationConfirm} disabled={catalogFiles.length === 0}>
                Confirm
              </button>
            </div>

            <div className="table-wrap catalog-table-wrap">
              {catalogLoading && !activeInformationFile ? (
                <div className="empty-state">Loading file catalog...</div>
              ) : activeInformationFile ? (
                <table className="catalog-table">
                  <thead>
                    <tr>
                      <th>Version</th>
                      <th>Previous Version</th>
                      <th>Hash</th>
                      <th>Miner Index</th>
                      <th>Size (MB)</th>
                      <th>Upload Timestamp</th>
                    </tr>
                  </thead>
                  <tbody>
                    {activeInformationFile.versions.map((version) => (
                      <tr key={version.cid}>
                        <td>{version.version_label}</td>
                        <td>{version.previous_version_label || "-"}</td>
                        <td className="hash-cell">{version.cid}</td>
                        <td>{version.miner_index || "-"}</td>
                        <td>{version.size_mb || "-"}</td>
                        <td>{formatTimestamp(version.uploaded_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="empty-state">
                  {catalogError || informationMessage || "Ready."}
                </div>
              )}
            </div>
          </div>
        </section>
      </main>

      {showFileBrowser ? (
        <FileBrowserModal
          onClose={() => setShowFileBrowser(false)}
          onSelect={handleUploadFileSelect}
        />
      ) : null}

      {previewOpen ? (
        <PreviewModal
          title={previewTitle}
          mode={previewMode}
          content={previewContent}
          resourceUrl={previewResourceUrl}
          onClose={() => setPreviewOpen(false)}
        />
      ) : null}
    </div>
  );
}
