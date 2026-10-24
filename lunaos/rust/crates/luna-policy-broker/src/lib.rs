use hmac::{Hmac, Mac};
use luna_contracts::{PolicyDecision, PolicyRequest, Principal, PrivacyClass};
use serde::{Deserialize, Serialize};
use sha2::Sha256;
use std::{
    io::{BufRead, BufReader, Read, Write},
    os::unix::net::UnixStream,
    path::{Path, PathBuf},
    time::{Duration, SystemTime, UNIX_EPOCH},
};
use subtle::ConstantTimeEq;
use thiserror::Error;

const MAX_RPC_BYTES: u64 = 8 * 1024;

#[derive(Debug, Error, PartialEq)]
pub enum BrokerError {
    #[error("owner authentication required")]
    OwnerAuthenticationRequired,
    #[error("action is not allowlisted")]
    ActionNotAllowlisted,
    #[error("payment refused because no budget is configured")]
    BudgetUnset,
    #[error("payment exceeds configured budget")]
    BudgetExceeded,
    #[error("sensitive action is not enabled")]
    SensitiveActionDenied,
    #[error("scoped action grant has expired")]
    GrantExpired,
    #[error("policy broker is unavailable or returned an invalid response")]
    ServiceUnavailable,
}

pub trait PolicyService: Send + Sync {
    fn authorize(&self, request: &PolicyRequest, goal_id: &str) -> Result<String, BrokerError>;
    fn verify(&self, goal_id: &str, action: &str, grant: &str) -> Result<(), BrokerError>;
}

#[derive(Clone)]
pub struct PolicyBroker {
    owner_token: Vec<u8>,
    budget_minor: Option<u64>,
}

impl PolicyBroker {
    pub fn new(owner_token: impl AsRef<[u8]>, budget_minor: Option<u64>) -> Self {
        Self {
            owner_token: owner_token.as_ref().to_vec(),
            budget_minor,
        }
    }

    pub fn evaluate(&self, request: &PolicyRequest) -> Result<PolicyDecision, BrokerError> {
        if request.principal != Principal::Owner || !self.owner_authenticated(&request.owner_token)
        {
            return Err(BrokerError::OwnerAuthenticationRequired);
        }
        if request.action.starts_with("payment.") {
            let Some(cost) = request.expected_cost_minor else {
                return Err(BrokerError::BudgetUnset);
            };
            let Some(budget) = self.budget_minor else {
                return Err(BrokerError::BudgetUnset);
            };
            if cost > budget {
                return Err(BrokerError::BudgetExceeded);
            }
            return Err(BrokerError::ActionNotAllowlisted);
        }
        if request.privacy_class != PrivacyClass::General {
            return Err(BrokerError::SensitiveActionDenied);
        }
        if request.action != "local.checksum" {
            return Err(BrokerError::ActionNotAllowlisted);
        }
        Ok(PolicyDecision {
            allowed: true,
            code: "OWNER_ALLOWLISTED_ACTION".into(),
        })
    }

    pub fn owner_authenticated(&self, token: &str) -> bool {
        token.as_bytes().ct_eq(&self.owner_token).into()
    }

    pub fn issue_grant(
        &self,
        request: &PolicyRequest,
        goal_id: &str,
    ) -> Result<String, BrokerError> {
        self.evaluate(request)?;
        let expires = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("system clock before epoch")
            .as_secs()
            + 3600;
        let mut mac =
            Hmac::<Sha256>::new_from_slice(&self.owner_token).expect("HMAC accepts any key size");
        mac.update(format!("{goal_id}:{}:{expires}", request.action).as_bytes());
        Ok(format!(
            "{expires}.{}",
            hex::encode(mac.finalize().into_bytes())
        ))
    }

    pub fn verify_grant(
        &self,
        goal_id: &str,
        action: &str,
        grant: &str,
    ) -> Result<(), BrokerError> {
        let (expires, signature) = grant
            .split_once('.')
            .ok_or(BrokerError::OwnerAuthenticationRequired)?;
        let expires: u64 = expires
            .parse()
            .map_err(|_| BrokerError::OwnerAuthenticationRequired)?;
        if SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("system clock before epoch")
            .as_secs()
            > expires
        {
            return Err(BrokerError::GrantExpired);
        }
        let mut mac =
            Hmac::<Sha256>::new_from_slice(&self.owner_token).expect("HMAC accepts any key size");
        mac.update(format!("{goal_id}:{action}:{expires}").as_bytes());
        let expected = mac.finalize().into_bytes();
        let supplied =
            hex::decode(signature).map_err(|_| BrokerError::OwnerAuthenticationRequired)?;
        if expected.as_slice().ct_eq(supplied.as_slice()).into() {
            Ok(())
        } else {
            Err(BrokerError::OwnerAuthenticationRequired)
        }
    }
}

impl PolicyService for PolicyBroker {
    fn authorize(&self, request: &PolicyRequest, goal_id: &str) -> Result<String, BrokerError> {
        self.issue_grant(request, goal_id)
    }

    fn verify(&self, goal_id: &str, action: &str, grant: &str) -> Result<(), BrokerError> {
        self.verify_grant(goal_id, action, grant)
    }
}

#[derive(Clone)]
pub struct UnixPolicyClient {
    socket: PathBuf,
    owner_token: String,
}

impl UnixPolicyClient {
    pub fn new(socket: impl AsRef<Path>, owner_token: impl Into<String>) -> Self {
        Self {
            socket: socket.as_ref().to_path_buf(),
            owner_token: owner_token.into(),
        }
    }

    fn call(&self, command: RpcCommand) -> Result<RpcResponse, BrokerError> {
        let mut stream =
            UnixStream::connect(&self.socket).map_err(|_| BrokerError::ServiceUnavailable)?;
        stream
            .set_read_timeout(Some(Duration::from_secs(2)))
            .map_err(|_| BrokerError::ServiceUnavailable)?;
        stream
            .set_write_timeout(Some(Duration::from_secs(2)))
            .map_err(|_| BrokerError::ServiceUnavailable)?;
        serde_json::to_writer(&mut stream, &command)
            .map_err(|_| BrokerError::ServiceUnavailable)?;
        stream
            .write_all(b"\n")
            .map_err(|_| BrokerError::ServiceUnavailable)?;
        stream
            .flush()
            .map_err(|_| BrokerError::ServiceUnavailable)?;
        let mut line = String::new();
        BufReader::new(stream)
            .take(MAX_RPC_BYTES)
            .read_line(&mut line)
            .map_err(|_| BrokerError::ServiceUnavailable)?;
        if line.is_empty() || line.len() as u64 >= MAX_RPC_BYTES {
            return Err(BrokerError::ServiceUnavailable);
        }
        serde_json::from_str::<RpcResponse>(&line).map_err(|_| BrokerError::ServiceUnavailable)
    }

    fn denied(response: RpcResponse) -> BrokerError {
        match response.code.as_str() {
            "OWNER_AUTH_REQUIRED" => BrokerError::OwnerAuthenticationRequired,
            "BUDGET_UNSET" => BrokerError::BudgetUnset,
            "BUDGET_EXCEEDED" => BrokerError::BudgetExceeded,
            "PRIVACY_SCOPE_DENIED" => BrokerError::SensitiveActionDenied,
            "GRANT_EXPIRED" => BrokerError::GrantExpired,
            "ACTION_DENIED" => BrokerError::ActionNotAllowlisted,
            _ => BrokerError::ServiceUnavailable,
        }
    }
}

impl PolicyService for UnixPolicyClient {
    fn authorize(&self, request: &PolicyRequest, goal_id: &str) -> Result<String, BrokerError> {
        let response = self.call(RpcCommand::Authorize {
            request: request.clone(),
            goal_id: goal_id.into(),
        })?;
        if !response.allowed {
            return Err(Self::denied(response));
        }
        response.grant.ok_or(BrokerError::ServiceUnavailable)
    }

    fn verify(&self, goal_id: &str, action: &str, grant: &str) -> Result<(), BrokerError> {
        let response = self.call(RpcCommand::Verify {
            owner_token: self.owner_token.clone(),
            goal_id: goal_id.into(),
            action: action.into(),
            grant: grant.into(),
        })?;
        if response.allowed {
            Ok(())
        } else {
            Err(Self::denied(response))
        }
    }
}

#[derive(Serialize, Deserialize)]
#[serde(tag = "op", rename_all = "snake_case", deny_unknown_fields)]
enum RpcCommand {
    Authorize {
        request: PolicyRequest,
        goal_id: String,
    },
    Verify {
        owner_token: String,
        goal_id: String,
        action: String,
        grant: String,
    },
}

#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct RpcResponse {
    allowed: bool,
    code: String,
    grant: Option<String>,
}

pub async fn serve_socket(path: impl AsRef<Path>, broker: PolicyBroker) -> std::io::Result<()> {
    use std::os::unix::fs::PermissionsExt;
    use tokio::{
        io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader as AsyncBufReader},
        net::UnixListener,
    };

    let path = path.as_ref();
    if path.exists() {
        return Err(std::io::Error::new(
            std::io::ErrorKind::AlreadyExists,
            "refusing to replace existing policy socket path",
        ));
    }
    let listener = UnixListener::bind(path)?;
    std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o600))?;
    loop {
        let (stream, _) = listener.accept().await?;
        let peer = stream.peer_cred()?;
        if peer.uid() != unsafe { libc::geteuid() } {
            continue;
        }
        let broker = broker.clone();
        tokio::spawn(async move {
            let (read_half, mut write_half) = stream.into_split();
            let reader = AsyncBufReader::new(read_half);
            let mut bytes = Vec::new();
            let read = reader
                .take(MAX_RPC_BYTES)
                .read_until(b'\n', &mut bytes)
                .await;
            let reply = match read {
                Ok(count) if count > 0 && (count as u64) < MAX_RPC_BYTES => {
                    match serde_json::from_slice::<RpcCommand>(&bytes) {
                        Ok(RpcCommand::Authorize { request, goal_id }) => {
                            match broker.issue_grant(&request, &goal_id) {
                                Ok(grant) => RpcResponse {
                                    allowed: true,
                                    code: "OWNER_SCOPED_GRANT".into(),
                                    grant: Some(grant),
                                },
                                Err(error) => RpcResponse {
                                    allowed: false,
                                    code: error_code(&error).into(),
                                    grant: None,
                                },
                            }
                        }
                        Ok(RpcCommand::Verify {
                            owner_token,
                            goal_id,
                            action,
                            grant,
                        }) => {
                            if !broker.owner_authenticated(&owner_token) {
                                RpcResponse {
                                    allowed: false,
                                    code: "OWNER_AUTH_REQUIRED".into(),
                                    grant: None,
                                }
                            } else {
                                match broker.verify_grant(&goal_id, &action, &grant) {
                                    Ok(()) => RpcResponse {
                                        allowed: true,
                                        code: "SCOPED_ACTION_ALLOWED".into(),
                                        grant: None,
                                    },
                                    Err(error) => RpcResponse {
                                        allowed: false,
                                        code: error_code(&error).into(),
                                        grant: None,
                                    },
                                }
                            }
                        }
                        Err(_) => RpcResponse {
                            allowed: false,
                            code: "INVALID_REQUEST".into(),
                            grant: None,
                        },
                    }
                }
                _ => RpcResponse {
                    allowed: false,
                    code: "INVALID_REQUEST".into(),
                    grant: None,
                },
            };
            if let Ok(mut response) = serde_json::to_vec(&reply) {
                response.push(b'\n');
                let _ = write_half.write_all(&response).await;
            }
        });
    }
}

fn error_code(error: &BrokerError) -> &'static str {
    match error {
        BrokerError::OwnerAuthenticationRequired => "OWNER_AUTH_REQUIRED",
        BrokerError::ActionNotAllowlisted => "ACTION_DENIED",
        BrokerError::BudgetUnset => "BUDGET_UNSET",
        BrokerError::BudgetExceeded => "BUDGET_EXCEEDED",
        BrokerError::SensitiveActionDenied => "PRIVACY_SCOPE_DENIED",
        BrokerError::GrantExpired => "GRANT_EXPIRED",
        BrokerError::ServiceUnavailable => "BROKER_UNAVAILABLE",
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use luna_contracts::{Principal, PrivacyClass};

    fn request(token: &str, principal: Principal, action: &str) -> PolicyRequest {
        PolicyRequest {
            owner_token: token.into(),
            principal,
            action: action.into(),
            privacy_class: PrivacyClass::General,
            expected_cost_minor: None,
            configured_budget_minor: None,
        }
    }

    #[test]
    fn only_authenticated_owner_gets_privileged_policy() {
        let broker = PolicyBroker::new("owner-secret", None);
        assert_eq!(
            broker.evaluate(&request("guest-secret", Principal::Guest, "local.checksum")),
            Err(BrokerError::OwnerAuthenticationRequired)
        );
        assert!(
            broker
                .evaluate(&request("owner-secret", Principal::Owner, "local.checksum"))
                .unwrap()
                .allowed
        );
    }

    #[test]
    fn payment_is_refused_when_budget_is_absent() {
        let broker = PolicyBroker::new("owner-secret", None);
        let mut req = request("owner-secret", Principal::Owner, "payment.purchase");
        req.expected_cost_minor = Some(100);
        assert_eq!(broker.evaluate(&req), Err(BrokerError::BudgetUnset));
    }

    #[test]
    fn unknown_action_is_denied() {
        let broker = PolicyBroker::new("owner-secret", None);
        assert_eq!(
            broker.evaluate(&request("owner-secret", Principal::Owner, "system.exec")),
            Err(BrokerError::ActionNotAllowlisted)
        );
    }

    #[test]
    fn grant_is_scoped_to_goal_and_action() {
        let broker = PolicyBroker::new("owner-secret", None);
        let req = request("owner-secret", Principal::Owner, "local.checksum");
        let grant = broker.issue_grant(&req, "goal-1").unwrap();
        assert!(
            broker
                .verify_grant("goal-1", "local.checksum", &grant)
                .is_ok()
        );
        assert!(
            broker
                .verify_grant("goal-2", "local.checksum", &grant)
                .is_err()
        );
        assert!(
            broker
                .verify_grant("goal-1", "system.exec", &grant)
                .is_err()
        );
    }
}
