use luna_contracts::{PolicyRequest, Principal, PrivacyClass};
use luna_policy_broker::{PolicyService, UnixPolicyClient};
use std::{
    fs,
    os::unix::fs::PermissionsExt,
    process::{Child, Command, Stdio},
    thread,
    time::{Duration, Instant},
};
use uuid::Uuid;

struct BrokerProcess(Child);

impl Drop for BrokerProcess {
    fn drop(&mut self) {
        let _ = self.0.kill();
        let _ = self.0.wait();
    }
}

#[test]
fn separate_broker_process_authenticates_owner_and_checks_scoped_grants() {
    let root = std::path::PathBuf::from("/tmp")
        .join(format!("lp-{}", &Uuid::new_v4().simple().to_string()[..8]));
    fs::create_dir(&root).unwrap();
    fs::set_permissions(&root, fs::Permissions::from_mode(0o700)).unwrap();
    let key_file = root.join("owner.key");
    let socket = root.join("policy.sock");
    let owner_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    fs::write(&key_file, owner_token).unwrap();
    fs::set_permissions(&key_file, fs::Permissions::from_mode(0o600)).unwrap();

    let child = Command::new(env!("CARGO_BIN_EXE_luna-policy-broker"))
        .env("LUNA_OWNER_TOKEN_FILE", &key_file)
        .env("LUNA_POLICY_SOCKET", &socket)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    let _broker = BrokerProcess(child);
    let deadline = Instant::now() + Duration::from_secs(3);
    while !socket.exists() && Instant::now() < deadline {
        thread::sleep(Duration::from_millis(10));
    }
    assert!(
        socket.exists(),
        "broker process did not create its Unix socket"
    );

    let client = UnixPolicyClient::new(&socket, owner_token);
    let authorized = PolicyRequest {
        owner_token: owner_token.into(),
        principal: Principal::Owner,
        action: "local.checksum".into(),
        privacy_class: PrivacyClass::General,
        expected_cost_minor: None,
        configured_budget_minor: None,
    };
    let grant = client.authorize(&authorized, "goal-process-test").unwrap();
    client
        .verify("goal-process-test", "local.checksum", &grant)
        .unwrap();
    assert!(
        client
            .verify("another-goal", "local.checksum", &grant)
            .is_err()
    );

    let guest = PolicyRequest {
        owner_token: "guest-token".into(),
        ..authorized.clone()
    };
    assert!(client.authorize(&guest, "goal-process-test").is_err());
    let payment = PolicyRequest {
        action: "payment.purchase".into(),
        expected_cost_minor: Some(1),
        ..authorized
    };
    assert!(client.authorize(&payment, "goal-payment-test").is_err());

    drop(_broker);
    fs::remove_dir_all(root).unwrap();
}
