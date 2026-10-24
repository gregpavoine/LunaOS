use sha2::{Digest, Sha256};
use thiserror::Error;

const PUBLIC_FIXTURE: &[u8] = b"LunaOS public checkpoint v1\n";

#[derive(Debug, Error, PartialEq, Eq)]
pub enum RunnerError {
    #[error("action is not enabled by the local runner")]
    ActionNotAllowed,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ArtifactEvidence {
    pub sha256: String,
    pub byte_count: u64,
    pub provenance: String,
}

pub trait ActionRunner: Send + Sync {
    fn execute(&self, action: &str) -> Result<ArtifactEvidence, RunnerError>;
}

/// Executes only an immutable public fixture action. This runner has no shell,
/// filesystem path, network or process execution interface.
pub struct LocalActionRunner;

impl ActionRunner for LocalActionRunner {
    fn execute(&self, action: &str) -> Result<ArtifactEvidence, RunnerError> {
        if action != "local.checksum" {
            return Err(RunnerError::ActionNotAllowed);
        }
        Ok(ArtifactEvidence {
            sha256: hex::encode(Sha256::digest(PUBLIC_FIXTURE)),
            byte_count: PUBLIC_FIXTURE.len() as u64,
            provenance: "local_public_fixture_v1".into(),
        })
    }
}
