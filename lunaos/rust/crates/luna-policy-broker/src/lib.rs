use hmac::{Hmac, Mac};
use luna_contracts::{PolicyDecision, PolicyRequest, Principal, PrivacyClass};
use sha2::Sha256;
use std::time::{SystemTime, UNIX_EPOCH};
use subtle::ConstantTimeEq;
use thiserror::Error;

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
        let token_matches: bool = request
            .owner_token
            .as_bytes()
            .ct_eq(&self.owner_token)
            .into();
        if request.principal != Principal::Owner || !token_matches {
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
        assert_eq!(
            broker
                .evaluate(&request("owner-secret", Principal::Owner, "local.checksum"))
                .unwrap()
                .allowed,
            true
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
