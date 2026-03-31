import React, { useEffect, useMemo, useState } from "react";
import sduLogo from "../sdu_logo.png";
import opendsnLogo from "../opendsn_logo.png";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080";
const PAGE_SIZE = 5;

function formatTimestamp(ts) {
  if (!ts) return "-";
  return ts.replace("T", " ").replace(/\+\d{2}:\d{2}$/, "");
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
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) {
    throw new Error(`Request failed: ${res.status}`);
  }
  return res.json();
}

async function postJSON(path, body) {
  const res = await fetch(`${API_BASE}${path}`, {
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

export default function App() {
  const [miners, setMiners] = useState([]);
  const [proofs, setProofs] = useState([]);
  const [minerPage, setMinerPage] = useState(1);
  const [proofPage, setProofPage] = useState(1);

  const [minersLoading, setMinersLoading] = useState(false);
  const [proofsLoading, setProofsLoading] = useState(false);

  const [importPath, setImportPath] = useState("");
  const [importLoading, setImportLoading] = useState(false);
  const [importMessage, setImportMessage] = useState("Ready.");
  const [importRoot, setImportRoot] = useState("");

  const [dealRoot, setDealRoot] = useState("");
  const [dealPreviousRoot, setDealPreviousRoot] = useState("");
  const [dealMiner, setDealMiner] = useState("");
  const [dealLoading, setDealLoading] = useState(false);
  const [dealMessage, setDealMessage] = useState("Ready.");

  const [retrieveRoot, setRetrieveRoot] = useState("");
  const [retrieveOutputName, setRetrieveOutputName] = useState("");
  const [retrieveLoading, setRetrieveLoading] = useState(false);
  const [retrieveMessage, setRetrieveMessage] = useState("Ready.");

  const [roots, setRoots] = useState([]);
  const [rootsLoading, setRootsLoading] = useState(false);

  const [versionRoot, setVersionRoot] = useState("");
  const [versions, setVersions] = useState([]);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [fileInfoMessage, setFileInfoMessage] = useState("Ready.");

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

  useEffect(() => {
    fetchMiners();
    fetchProofs();

    const timer = setInterval(() => {
      fetchMiners();
      fetchProofs();
    }, 30000);

    return () => clearInterval(timer);
  }, []);

  const minerPagination = useMemo(
    () => paginate(miners, minerPage, PAGE_SIZE),
    [miners, minerPage]
  );

  const proofPagination = useMemo(
    () => paginate(proofs, proofPage, PAGE_SIZE),
    [proofs, proofPage]
  );

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

  async function handleImport() {
    if (!importPath.trim()) {
      setImportMessage("Please enter a file path.");
      return;
    }

    try {
      setImportLoading(true);
      setImportMessage("Importing...");
      const data = await postJSON("/api/client/import", {
        path: importPath.trim(),
      });

      if (data.success) {
        setImportRoot(data.root || "");
        setDealRoot(data.root || "");
        setImportMessage(`Import finished. Root: ${data.root || "-"}`);
      } else {
        setImportMessage(data.message || "Import failed.");
      }
    } catch (err) {
      setImportMessage(`Request failed: ${err.message}`);
    } finally {
      setImportLoading(false);
    }
  }

  async function handleDeal() {
    if (!dealRoot.trim()) {
      setDealMessage("Root is required.");
      return;
    }
    if (!dealMiner.trim()) {
      setDealMessage("Miner index is required.");
      return;
    }

    try {
      setDealLoading(true);
      setDealMessage("Uploading...");
      const body = {
        root: dealRoot.trim(),
        miner: dealMiner.trim(),
      };
      if (dealPreviousRoot.trim()) {
        body.previous_root = dealPreviousRoot.trim();
      }

      const data = await postJSON("/api/client/deal", body);

      if (data.success) {
        setDealMessage("Deal finished.");
      } else {
        setDealMessage("Deal failed.");
      }
    } catch (err) {
      setDealMessage(`Request failed: ${err.message}`);
    } finally {
      setDealLoading(false);
    }
  }

  async function handleRetrieve() {
    if (!retrieveRoot.trim()) {
      setRetrieveMessage("Root / head is required.");
      return;
    }
    if (!retrieveOutputName.trim()) {
      setRetrieveMessage("Output file name is required.");
      return;
    }

    try {
      setRetrieveLoading(true);
      setRetrieveMessage("Retrieving...");
      const data = await postJSON("/api/client/retrieve-version", {
        root: retrieveRoot.trim(),
        output_name: retrieveOutputName.trim(),
      });

      if (data.success) {
        setRetrieveMessage(
          data.output_path
            ? `Retrieve finished: ${data.output_path}`
            : "Retrieve finished."
        );
      } else {
        setRetrieveMessage(data.message || "Retrieve failed.");
      }
    } catch (err) {
      setRetrieveMessage(`Request failed: ${err.message}`);
    } finally {
      setRetrieveLoading(false);
    }
  }

  async function handleShowRoots() {
    try {
      setRootsLoading(true);
      setFileInfoMessage("");
      setVersions([]);
      const data = await getJSON("/api/client/roots");
      setRoots(data.roots || []);
      if (!data.roots || data.roots.length === 0) {
        setFileInfoMessage("No roots found.");
      }
    } catch (err) {
      setFileInfoMessage(err.message);
      setRoots([]);
    } finally {
      setRootsLoading(false);
    }
  }

  async function handleShowVersions() {
    if (!versionRoot.trim()) {
      setFileInfoMessage("Please enter a file root.");
      return;
    }

    try {
      setVersionsLoading(true);
      setFileInfoMessage("");
      setRoots([]);
      const data = await getJSON(
        `/api/client/versions?root=${encodeURIComponent(versionRoot.trim())}`
      );
      setVersions(data.versions || []);
      if (!data.versions || data.versions.length === 0) {
        setFileInfoMessage("No versions found.");
      }
    } catch (err) {
      setFileInfoMessage(err.message);
      setVersions([]);
    } finally {
      setVersionsLoading(false);
    }
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
              onPrev={() => setMinerPage((p) => Math.max(1, p - 1))}
              onNext={() =>
                setMinerPage((p) => Math.min(minerPagination.totalPages, p + 1))
              }
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
              onPrev={() => setProofPage((p) => Math.max(1, p - 1))}
              onNext={() =>
                setProofPage((p) => Math.min(proofPagination.totalPages, p + 1))
              }
            />
          </div>
        </section>

        <section className="action-section">
          <div className="card">
            <SectionTitle>File Upload</SectionTitle>

            <label className="field-label">File Import Path</label>
            <input
              type="text"
              placeholder="Enter file path"
              value={importPath}
              onChange={(e) => setImportPath(e.target.value)}
            />
            <button onClick={handleImport} disabled={importLoading}>
              {importLoading ? "Importing..." : "Import File"}
            </button>

            <label className="field-label">File Root</label>
            <input
              type="text"
              placeholder="Enter file root"
              value={dealRoot}
              onChange={(e) => setDealRoot(e.target.value)}
            />

            <label className="field-label">Previous Version Root</label>
            <input
              type="text"
              placeholder="Optional previous root"
              value={dealPreviousRoot}
              onChange={(e) => setDealPreviousRoot(e.target.value)}
            />

            <label className="field-label">Miner Index</label>
            <input
              type="text"
              placeholder="Enter miner index, e.g. t01000"
              value={dealMiner}
              onChange={(e) => setDealMiner(e.target.value)}
            />

            <button onClick={handleDeal} disabled={dealLoading}>
              {dealLoading ? "Uploading..." : "Upload File"}
            </button>

            <div className="message-box upload-message-box">
              {importMessage === "Ready." && dealMessage === "Ready." ? (
                <div className="empty-state">Ready.</div>
              ) : (
                <>
                  {importMessage !== "Ready." && <div>{importMessage}</div>}
                  {dealMessage !== "Ready." && <div>{dealMessage}</div>}
                </>
              )}
            </div>
          </div>

          <div className="card">
            <SectionTitle>File Retrieval</SectionTitle>

            <label className="field-label">Root / Head</label>
            <input
              type="text"
              placeholder="Enter root or head"
              value={retrieveRoot}
              onChange={(e) => setRetrieveRoot(e.target.value)}
            />

            <label className="field-label">Output File Name</label>
            <input
              type="text"
              placeholder="Enter output file name"
              value={retrieveOutputName}
              onChange={(e) => setRetrieveOutputName(e.target.value)}
            />

            <button onClick={handleRetrieve} disabled={retrieveLoading}>
              {retrieveLoading ? "Retrieving..." : "Retrieve File"}
            </button>

            <div className="message-box">
              <div className={retrieveMessage === "Ready." ? "empty-state" : ""}>
                {retrieveMessage}
              </div>
            </div>
          </div>

          <div className="card">
            <SectionTitle>File Information</SectionTitle>

            <button onClick={handleShowRoots} disabled={rootsLoading}>
              {rootsLoading ? "Loading Roots..." : "Show File Roots"}
            </button>

            <input
              type="text"
              placeholder="Enter file root"
              value={versionRoot}
              onChange={(e) => setVersionRoot(e.target.value)}
            />

            <button onClick={handleShowVersions} disabled={versionsLoading}>
              {versionsLoading ? "Loading Versions..." : "Show All Heads"}
            </button>

            <div className="file-box">
              {roots.length > 0 &&
                roots.map((item, index) => (
                  <div key={item.root} className="file-item file-record">
                    <div className="file-record-label">ROOT {index + 1}</div>
                    <div className="file-record-value">{item.root}</div>
                  </div>
                ))}

              {versions.length > 0 &&
                versions.map((item) => (
                  <div key={`${item.label}-${item.cid}`} className="file-item file-record">
                    <div className="file-record-label">{item.label}</div>
                    <div className="file-record-value">{item.cid}</div>
                  </div>
                ))}

              {roots.length === 0 && versions.length === 0 && (
                <div className="empty-state">{fileInfoMessage}</div>
              )}
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
